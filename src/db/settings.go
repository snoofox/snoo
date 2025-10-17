package db

import "gorm.io/gorm"

func GetSetting(db *gorm.DB, key string) (string, error) {
	var setting Setting
	result := db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return "", result.Error
	}
	return setting.Value, nil
}

func SetSetting(db *gorm.DB, key, value string) error {
	var setting Setting
	result := db.Where("key = ?", key).First(&setting)

	if result.Error == gorm.ErrRecordNotFound {
		setting = Setting{Key: key, Value: value}
		return db.Create(&setting).Error
	}

	setting.Value = value
	return db.Save(&setting).Error
}
