package handler

import (
	"Permia/pkg/response"
	"Permia/core-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc *service.ProductService
}

func NewProductHandler(productSvc *service.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

// ListProducts لیست کامل محصولات دسته‌بندی شده
func (h *ProductHandler) ListProducts(c *gin.Context) {
	catalog, err := h.productSvc.GetCatalog(c)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, catalog, "Products retrieved successfully")
}