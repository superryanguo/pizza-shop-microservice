package handlers

import (
	"net/http"

	"github.com/superryanguo/pizza/clients"
	"github.com/superryanguo/pizza/pizza"
	"github.com/superryanguo/pizza/pizza/services"
	"github.com/superryanguo/pizza/users/utils"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService     pizza.OrderService
	orderItemService services.OrderItemService
	utilityService   utils.UtilityService
	prometheusClient clients.PrometheusClient
}

func NewOrderHandler(orderService pizza.OrderService, orderItemService services.OrderItemService, utilityService utils.UtilityService, prometheusClient clients.PrometheusClient) *OrderHandler {
	return &OrderHandler{
		orderService:     orderService,
		orderItemService: orderItemService,
		utilityService:   utilityService,
		prometheusClient: prometheusClient,
	}

}

func (o OrderHandler) CreateOrder(c *gin.Context) {
	userID := o.utilityService.GetUserFromContext(c)
	if userID == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err := o.orderService.CreateOrder(c, userID)
	if err != nil {
		if err.Error() == "no-cart" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	go o.prometheusClient.RecordTotalNumberOfOrders()
	c.Status(http.StatusCreated)
}
