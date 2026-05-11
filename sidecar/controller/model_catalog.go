package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	sidecarModel "github.com/QuantumNous/new-api/sidecar/model"
	"github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
)

// --- Public API ---

func GetModelCatalog(c *gin.Context) {
	vendor := c.Query("vendor")
	capabilities := c.Query("capabilities")
	models, err := service.GetModelCatalog(vendor, capabilities)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, models)
}

func GetModelCatalogByName(c *gin.Context) {
	modelName := c.Param("model_name")
	cm, err := service.GetModelCatalogByName(modelName)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	if cm == nil {
		common.ApiErrorI18n(c, i18n.MsgNotFound)
		return
	}
	common.ApiSuccess(c, cm)
}

// --- Admin API ---

type CreateModelSpecRequest struct {
	ModelName       string   `json:"model_name" binding:"required"`
	ContextLength   int      `json:"context_length"`
	MaxOutputTokens int      `json:"max_output_tokens"`
	Capabilities    []string `json:"capabilities"`
	Description     string   `json:"description"`
	Icon            string   `json:"icon"`
	ReleaseDate     string   `json:"release_date"`
	KnowledgeCutoff string   `json:"knowledge_cutoff"`
	ParameterCount  string   `json:"parameter_count"`
	Status          int      `json:"status"`
}

type UpdateModelSpecRequest struct {
	Id              int      `json:"id" binding:"required"`
	ModelName       string   `json:"model_name" binding:"required"`
	ContextLength   int      `json:"context_length"`
	MaxOutputTokens int      `json:"max_output_tokens"`
	Capabilities    []string `json:"capabilities"`
	Description     string   `json:"description"`
	Icon            string   `json:"icon"`
	ReleaseDate     string   `json:"release_date"`
	KnowledgeCutoff string   `json:"knowledge_cutoff"`
	ParameterCount  string   `json:"parameter_count"`
	Status          int      `json:"status"`
}

func AdminListModelSpecs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")
	specs, total, err := sidecarModel.SearchModelSpecs(keyword, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(specs)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetModelSpec(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	spec, err := sidecarModel.GetModelSpecByID(id)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, spec)
}

func AdminCreateModelSpec(c *gin.Context) {
	var req CreateModelSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	if dup, _ := sidecarModel.IsModelSpecNameDuplicated(0, req.ModelName); dup {
		common.ApiErrorI18n(c, i18n.MsgAlreadyExists)
		return
	}
	caps, _ := common.Marshal(req.Capabilities)
	spec := &sidecarModel.ModelSpec{
		ModelName:       req.ModelName,
		ContextLength:   req.ContextLength,
		MaxOutputTokens: req.MaxOutputTokens,
		Capabilities:    string(caps),
		Description:     req.Description,
		Icon:            req.Icon,
		ReleaseDate:     req.ReleaseDate,
		KnowledgeCutoff: req.KnowledgeCutoff,
		ParameterCount:  req.ParameterCount,
		Status:          req.Status,
	}
	if spec.Status == 0 {
		spec.Status = 1
	}
	if err := spec.Insert(); err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, spec)
}

func AdminUpdateModelSpec(c *gin.Context) {
	var req UpdateModelSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	if dup, _ := sidecarModel.IsModelSpecNameDuplicated(req.Id, req.ModelName); dup {
		common.ApiErrorI18n(c, i18n.MsgAlreadyExists)
		return
	}
	caps, _ := common.Marshal(req.Capabilities)
	spec := &sidecarModel.ModelSpec{
		Id:              req.Id,
		ModelName:       req.ModelName,
		ContextLength:   req.ContextLength,
		MaxOutputTokens: req.MaxOutputTokens,
		Capabilities:    string(caps),
		Description:     req.Description,
		Icon:            req.Icon,
		ReleaseDate:     req.ReleaseDate,
		KnowledgeCutoff: req.KnowledgeCutoff,
		ParameterCount:  req.ParameterCount,
		Status:          req.Status,
	}
	if err := spec.Update(); err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, spec)
}

func AdminDeleteModelSpec(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	if err := sidecarModel.DeleteModelSpec(id); err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	common.ApiSuccess(c, nil)
}
