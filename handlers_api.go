package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// REST API 处理函数

// respondError 返回错误响应
func respondError(c *gin.Context, statusCode int, code, message string, details any) {
	response := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}

	logrus.Errorf("%s %s %d", c.Request.Method, c.Request.URL.Path, statusCode)
	c.JSON(statusCode, response)
}

// respondSuccess 返回成功响应
func respondSuccess(c *gin.Context, data any, message string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	logrus.Infof("%s %s %d", c.Request.Method, c.Request.URL.Path, http.StatusOK)
	c.JSON(http.StatusOK, response)
}

// checkLoginStatusHandler 检查登录状态
func (s *AppServer) checkLoginStatusHandler(c *gin.Context) {
	status, err := s.dianshuService.CheckLoginStatus(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "STATUS_CHECK_FAILED", "检查登录状态失败", err.Error())
		return
	}
	respondSuccess(c, status, "检查登录状态成功")
}

// getLoginQRCodeHandler 获取登录二维码
func (s *AppServer) getLoginQRCodeHandler(c *gin.Context) {
	result, err := s.dianshuService.GetLoginQRCode(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "QRCODE_FAILED", "获取登录二维码失败", err.Error())
		return
	}
	respondSuccess(c, result, "获取登录二维码成功")
}

// deleteCookiesHandler 删除 cookies
func (s *AppServer) deleteCookiesHandler(c *gin.Context) {
	result, err := s.dianshuService.DeleteCookies(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DELETE_COOKIES_FAILED", "删除 cookies 失败", err.Error())
		return
	}
	respondSuccess(c, result, "删除 cookies 成功")
}

// listOrdersHandler 查询订单列表
func (s *AppServer) listOrdersHandler(c *gin.Context) {
	var req OrderQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数错误", err.Error())
		return
	}

	result, err := s.dianshuService.QueryOrders(c.Request.Context(), req.OrderType, req.OrderCode)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "QUERY_ORDERS_FAILED", "查询订单失败", err.Error())
		return
	}

	respondSuccess(c, result, "查询订单成功")
}

// dataSearchHandler 数据查询。
func (s *AppServer) xhsSearchHandler(c *gin.Context) {
	var request DataSearchArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数错误", err.Error())
		return
	}

	result, err := s.dianshuService.XhsSearch(c.Request.Context(), request)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DATA_SEARCH_FAILED", "数据查询失败", err.Error())
		return
	}
	respondSuccess(c, result, "数据查询成功")
}

// datasetSearchHandler 典枢数据集搜索。
func (s *AppServer) datasetSearchHandler(c *gin.Context) {
	var request DatasetSearchArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数错误", err.Error())
		return
	}

	result, err := s.dianshuService.SearchDatasets(c.Request.Context(), request.Keyword, request.PageNo, request.PageSize)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DATASET_SEARCH_FAILED", "数据集搜索失败", err.Error())
		return
	}
	respondSuccess(c, result, "数据集搜索成功")
}
