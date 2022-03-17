package implementation

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/superryanguo/kitchen/clients"
	"github.com/superryanguo/kitchen/cooks"
	"github.com/superryanguo/kitchen/inmemorydb"
	"github.com/superryanguo/kitchen/processes"
	"github.com/superryanguo/kitchen/queue"
	"github.com/superryanguo/kitchen/shared"
)

type processorderimplementation struct {
	cookservice       cooks.Service
	service           processes.OrderProcessUpdateService
	orderQueueService inmemorydb.OrderRequestInMemoryService
	messageQueueRepo  queue.QueueRepository
	prometheusClient  clients.PrometheusClient
}

func NewProcessOrderImplementationService(cs cooks.Service, service processes.OrderProcessUpdateService, oq inmemorydb.OrderRequestInMemoryService, mrp queue.QueueRepository, prometheusClient clients.PrometheusClient) processes.OrderProcessService {
	return &processorderimplementation{
		cookservice:       cs,
		service:           service,
		orderQueueService: oq,
		messageQueueRepo:  mrp,
		prometheusClient:  prometheusClient,
	}
}

func (poi processorderimplementation) ProcessOrder(ctx context.Context, orderRequest queue.OrderQueueRequest, cookID int, updateStatus bool) {
	/*
		1) Mark the cook as not available.
		2) Loop through each order details.
		3) Process each pizza.
		4) Mark the details in DB.
		5) If all the items are processed sucessfully  mark the order status in DB.

	*/

	var err error
	go func() {
		err = poi.cookservice.UpdateCookStatus(ctx, cookID, 0)
		if err != nil {
			glog.Errorf("Unable to update the status of cook %s", err)
			// Make
		}
		for _, item := range orderRequest.Details {
			/*
				Just sleeping for 60 seconds to  simulate it as a expensive
				process
			*/
			for i := 0; i < item.Quantity; i++ {
				time.Sleep(shared.TimeToMakePizza)
				glog.Info("Pizza %s is ready...", item.PizzaID)
				poi.service.MarkOrderItemComplete(ctx, item.PizzaID, orderRequest.OrderUUID)
			}
		}

		poi.service.MarkOrderComplete(ctx, orderRequest.OrderUUID, cookID)
		go poi.prometheusClient.RecordCompletedOrders()
		go poi.messageQueueRepo.PublishOrderStatus(ctx, orderRequest.OrderUUID, shared.OrderStatusDelivered)
		//Update the status to cook service via message queue

		// Free the cook whether the order fails or not
		/*
			See if there is any order in queue if there is orders
			in queue assign this cook to that order

		*/
		order, err := poi.orderQueueService.GetOrder(ctx)
		if err != nil {
			glog.Error("Error in getting the first order...", err)
			poi.cookservice.UpdateCookStatus(ctx, cookID, 1)
			return
		}
		if order != nil {
			glog.Info("Order is in the DB so assigning this cook to that order")
			poi.ProcessOrder(ctx, *order, cookID, false)
			return
		} else {
			poi.cookservice.UpdateCookStatus(ctx, cookID, 1)
		}
	}()

}
