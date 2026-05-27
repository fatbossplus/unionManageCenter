// Package worker 后台工作器：定时调度采集任务 + AI 流水线处理
package worker

import (
	"log"
	"time"

	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/cms/internal/service"
	"unionManageCenter/pkg/database"
)

// Scheduler 任务调度器（每分钟扫一次到期任务）
type Scheduler struct {
	db         *gorm.DB
	scraperSvc *service.ScraperService
	stop       chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		db:         database.Get(),
		scraperSvc: service.NewScraperService(),
		stop:       make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("[CMS Scheduler] 启动，间隔 1 分钟")
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.tick()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() { close(s.stop) }

func (s *Scheduler) tick() {
	var tasks []model.ScrapeTask
	s.db.Where(
		"deleted_at IS NULL AND status = 1 AND cron_expr != '' AND (next_run_at IS NULL OR next_run_at <= ?)",
		time.Now(),
	).Find(&tasks)

	for _, task := range tasks {
		t := task
		go func() {
			count, err := s.scraperSvc.RunTask(&t)
			if err != nil {
				log.Printf("[Scheduler] 任务[%d]%s 执行失败: %v", t.ID, t.TaskName, err)
				s.db.Model(&t).Update("last_error", err.Error())
			} else {
				log.Printf("[Scheduler] 任务[%d]%s 完成，新增 %d 条内容", t.ID, t.TaskName, count)
			}
			// 计算下次执行时间（简化版：每小时/每天）
			next := calcNext(t.CronExpr)
			s.db.Model(&t).Update("next_run_at", next)
		}()
	}
}

// calcNext 简化版 cron 计算（支持 @hourly / @daily / @weekly）
func calcNext(expr string) time.Time {
	switch expr {
	case "@hourly":
		return time.Now().Add(time.Hour)
	case "@daily":
		return time.Now().Add(24 * time.Hour)
	case "@weekly":
		return time.Now().Add(7 * 24 * time.Hour)
	default:
		return time.Now().Add(time.Hour) // 默认1小时
	}
}

// Processor AI 流水线处理器（每30秒处理一批 pending 内容）
type Processor struct {
	db   *gorm.DB
	stop chan struct{}
}

func NewProcessor() *Processor {
	return &Processor{db: database.Get(), stop: make(chan struct{})}
}

func (p *Processor) Start() {
	log.Println("[CMS Processor] 启动，间隔 30 秒")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.processBatch()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *Processor) Stop() { close(p.stop) }

func (p *Processor) processBatch() {
	var contents []model.RawContent
	p.db.Where("proc_status = ?", model.ProcStatusPending).
		Order("id ASC").
		Limit(5). // 每批处理5条，避免大量并发
		Find(&contents)

	for _, raw := range contents {
		r := raw
		go func() {
			// 从任务配置读取目标平台
			var task model.ScrapeTask
			if err := p.db.First(&task, r.TaskID).Error; err != nil {
				log.Printf("[Processor] 找不到任务[%d]，跳过内容[%d]", r.TaskID, r.ID)
				p.db.Model(&r).Update("proc_status", model.ProcStatusSkipped)
				return
			}

			pipeline := service.NewPipelineService("ollama_free", []string{"wordlist_free"})
			finalText, err := pipeline.ProcessRaw(&r, task.TargetPlatform)
			if err != nil {
				log.Printf("[Processor] 内容[%d]处理失败: %v", r.ID, err)
				return
			}
			if finalText != "" {
				p.db.Create(&model.PublishTask{
					OrgID:          r.OrgID,
					RawID:          r.ID,
					AccountID:      task.AccountID,
					TargetPlatform: task.TargetPlatform,
					FinalTitle:     r.Title,
					FinalText:      finalText,
					Status:         model.PublishStatusDraft,
				})
			}
		}()
	}
}
