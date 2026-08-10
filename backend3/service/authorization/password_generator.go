package authorization

import (
	"crypto/rand"
	"math/big"
)

// PasswordConfig 密码配置
type PasswordConfig struct {
	Length      int    // 密码长度
	UpperCount  int    // 大写字母数量
	LowerCount  int    // 小写字母数量
	DigitCount  int    // 数字数量
	SpecialChar string // 特殊字符
}

// DefaultPasswordConfig 默认密码配置
var DefaultPasswordConfig = PasswordConfig{
	Length:      12,
	UpperCount:  2,
	LowerCount:  4,
	DigitCount:  4,
	SpecialChar: "@#$%",
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(config PasswordConfig) (string, error) {
	if config.Length < config.UpperCount+config.LowerCount+config.DigitCount {
		config.Length = config.UpperCount + config.LowerCount + config.DigitCount
	}

	password := make([]byte, config.Length)
	position := 0

	// 添加大写字母
	upperChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < config.UpperCount; i++ {
		char, err := selectRandomChar(upperChars)
		if err != nil {
			return "", err
		}
		password[position] = char
		position++
	}

	// 添加小写字母
	lowerChars := "abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < config.LowerCount; i++ {
		char, err := selectRandomChar(lowerChars)
		if err != nil {
			return "", err
		}
		password[position] = char
		position++
	}

	// 添加数字
	digitChars := "0123456789"
	for i := 0; i < config.DigitCount; i++ {
		char, err := selectRandomChar(digitChars)
		if err != nil {
			return "", err
		}
		password[position] = char
		position++
	}

	// 填充剩余位置（使用所有字符类型）
	allChars := upperChars + lowerChars + digitChars + config.SpecialChar
	for i := position; i < config.Length; i++ {
		char, err := selectRandomChar(allChars)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// 随机打乱密码字符顺序
	shuffledPassword, err := shufflePassword(password)
	if err != nil {
		return "", err
	}

	return string(shuffledPassword), nil
}

// selectRandomChar 从字符集中随机选择一个字符
func selectRandomChar(chars string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, err
	}
	return chars[n.Int64()], nil
}

// shufflePassword 打乱密码字符顺序
func shufflePassword(password []byte) ([]byte, error) {
	for i := len(password) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
	return password, nil
}
