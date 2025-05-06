package config

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Мокаем функцию ParseSize
func mockParseSize(size string) (int64, error) {
	if size == "1MB" {
		return 1024 * 1024, nil // 1MB in bytes
	}
	return 0, errors.New("invalid size")
}

func TestNewCommunityConfig_Success(t *testing.T) {
	// Создадим временный конфигурационный файл для теста
	configContent := `
community_name_min_length = 3
community_name_max_length = 20
community_description_max_length = 100
community_avatar_max_size = "1MB"
`

	// Создаем временный файл
	tempFile, err := ioutil.TempFile("", "config_test_*.toml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(configContent)
	assert.NoError(t, err)

	// Тестируем успешную загрузку конфигурации
	cfg, err := NewCommunityConfig(tempFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Проверка значений
	assert.Equal(t, 3, cfg.CommunityNameMinLength)
	assert.Equal(t, 20, cfg.CommunityNameMaxLength)
	assert.Equal(t, 100, cfg.CommunityDescriptionMaxLength)
	assert.Equal(t, int64(1024*1024), cfg.CommunityAvatarMaxSize)
}

func TestNewCommunityConfig_FileNotFound(t *testing.T) {
	// Тестируем ситуацию с несуществующим файлом
	cfg, err := NewCommunityConfig("non_existing_file.toml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unable to parse validation config from file")
}
