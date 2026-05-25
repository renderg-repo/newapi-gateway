package model

import (
	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

func Migrate() error {
	return DB.AutoMigrate(&ModelSpec{})
}

type ModelSpec struct {
	Id              int            `json:"id"`
	ModelName       string         `json:"model_name" gorm:"size:128;not null;uniqueIndex:uk_model_spec_name_delete_at,priority:1"`
	ContextLength   int            `json:"context_length"`
	MaxOutputTokens int            `json:"max_output_tokens"`
	Capabilities    string         `json:"capabilities,omitempty" gorm:"type:text"`
	Description     string         `json:"description,omitempty" gorm:"type:text"`
	Icon            string         `json:"icon,omitempty" gorm:"size:varchar(128)"`
	ReleaseDate     string         `json:"release_date,omitempty" gorm:"size:32"`
	KnowledgeCutoff string         `json:"knowledge_cutoff,omitempty" gorm:"size:32"`
	ParameterCount  string         `json:"parameter_count,omitempty" gorm:"size:32"`
	Status          int            `json:"status" gorm:"default:1"`
	CreatedTime     int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime     int64          `json:"updated_time" gorm:"bigint"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index;uniqueIndex:uk_model_spec_name_delete_at,priority:2"`
}

func (s *ModelSpec) Insert() error {
	now := common.GetTimestamp()
	s.CreatedTime = now
	s.UpdatedTime = now
	originalStatus := s.Status
	if err := DB.Create(s).Error; err != nil {
		return err
	}
	return DB.Model(&ModelSpec{}).Where("id = ?", s.Id).Update("status", originalStatus).Error
}

func (s *ModelSpec) AfterFind() error {
	return nil
}

func (s *ModelSpec) Update() error {
	s.UpdatedTime = common.GetTimestamp()
	return DB.Model(&ModelSpec{}).Where("id = ?", s.Id).Updates(map[string]interface{}{
		"model_name":        s.ModelName,
		"context_length":    s.ContextLength,
		"max_output_tokens": s.MaxOutputTokens,
		"capabilities":      s.Capabilities,
		"description":       s.Description,
		"icon":              s.Icon,
		"release_date":      s.ReleaseDate,
		"knowledge_cutoff":  s.KnowledgeCutoff,
		"parameter_count":   s.ParameterCount,
		"status":            s.Status,
		"updated_time":      s.UpdatedTime,
	}).Error
}

func GetAllModelSpecs(startIdx int, pageSize int) ([]*ModelSpec, error) {
	var specs []*ModelSpec
	err := DB.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&specs).Error
	return specs, err
}

func SearchModelSpecs(keyword string, startIdx int, pageSize int) ([]*ModelSpec, int64, error) {
	var specs []*ModelSpec
	var total int64
	query := DB.Model(&ModelSpec{})
	if keyword != "" {
		query = query.Where("model_name LIKE ?", "%"+keyword+"%")
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&specs).Error
	return specs, total, err
}

func GetModelSpecByModelName(modelName string) (*ModelSpec, error) {
	var spec ModelSpec
	err := DB.Where("model_name = ? AND status = ?", modelName, 1).First(&spec).Error
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func GetModelSpecByID(id int) (*ModelSpec, error) {
	var spec ModelSpec
	err := DB.First(&spec, id).Error
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func DeleteModelSpec(id int) error {
	return DB.Delete(&ModelSpec{}, id).Error
}

func IsModelSpecNameDuplicated(id int, name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	var cnt int64
	err := DB.Model(&ModelSpec{}).Where("model_name = ? AND id <> ?", name, id).Count(&cnt).Error
	return cnt > 0, err
}
