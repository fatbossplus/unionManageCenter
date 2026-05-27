package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/pkg/database"
)

// masterKey 全局主密钥（32字节 AES-256）
// 从环境变量 CMS_MASTER_KEY 读取（64位hex字符串）
// 未设置则用默认值（仅开发环境，生产必须设置）
var (
	masterKey     []byte
	masterKeyOnce sync.Once
)

func getMasterKey() ([]byte, error) {
	var keyErr error
	masterKeyOnce.Do(func() {
		hex64 := os.Getenv("CMS_MASTER_KEY")
		if hex64 == "" {
			// 开发默认值，32字节
			hex64 = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		}
		key, err := hex.DecodeString(hex64)
		if err != nil || len(key) != 32 {
			keyErr = errors.New("CMS_MASTER_KEY 必须是64位十六进制字符串（32字节）")
			return
		}
		masterKey = key
	})
	return masterKey, keyErr
}

// deriveKey 用主密钥+账号ID派生账号专属密钥（防止单个密钥泄露影响全部账号）
func deriveKey(accountID uint64) ([]byte, error) {
	kek, err := getMasterKey()
	if err != nil {
		return nil, err
	}
	// HKDF 简化版：SHA256(masterKey || accountID)
	h := sha256.New()
	h.Write(kek)
	idBuf := make([]byte, 8)
	idBuf[0] = byte(accountID >> 56)
	idBuf[1] = byte(accountID >> 48)
	idBuf[2] = byte(accountID >> 40)
	idBuf[3] = byte(accountID >> 32)
	idBuf[4] = byte(accountID >> 24)
	idBuf[5] = byte(accountID >> 16)
	idBuf[6] = byte(accountID >> 8)
	idBuf[7] = byte(accountID)
	h.Write(idBuf)
	return h.Sum(nil), nil // SHA256 = 32 bytes = 256 bit AES key
}

// EncryptCred 用账号专属密钥加密凭证明文，返回 (密文Base64, IV Base64, error)
func EncryptCred(accountID uint64, plaintext string) (string, string, error) {
	key, err := deriveKey(accountID)
	if err != nil {
		return "", "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}
	cipherBytes := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(cipherBytes),
		base64.StdEncoding.EncodeToString(nonce),
		nil
}

// DecryptCred 解密凭证密文，返回明文
func DecryptCred(accountID uint64, cipherB64, ivB64 string) (string, error) {
	key, err := deriveKey(accountID)
	if err != nil {
		return "", err
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", errors.New("凭证解密失败：密文损坏或密钥不匹配")
	}
	return string(plain), nil
}

// CredentialService 凭证服务（含审计）
type CredentialService struct {
	db *gorm.DB
}

func NewCredentialService() *CredentialService {
	return &CredentialService{db: database.Get()}
}

// WriteAudit 记录审计日志（异步，不阻塞主流程）
func (s *CredentialService) WriteAudit(accountID, operatorID uint64, action, reason, ip string) {
	go func() {
		s.db.Create(&model.CredentialAudit{
			AccountID:  accountID,
			OperatorID: operatorID,
			Action:     action,
			Reason:     reason,
			IP:         ip,
			CreatedAt:  time.Now(),
		})
	}()
}

// GetDecrypted 获取账号解密凭证（记录审计）
func (s *CredentialService) GetDecrypted(accountID, operatorID uint64, reason, ip string) (string, error) {
	var acc model.PlatformAccount
	if err := s.db.First(&acc, accountID).Error; err != nil {
		return "", err
	}
	s.WriteAudit(accountID, operatorID, "read", reason, ip)
	acc.LastUsedAt = func() *time.Time { t := time.Now(); return &t }()
	s.db.Model(&acc).Update("last_used_at", acc.LastUsedAt)

	return DecryptCred(accountID, acc.CredCipher, acc.CredIV)
}
