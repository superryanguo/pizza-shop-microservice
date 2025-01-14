package implementation

import (
	"context"
	"errors"
	"github.com/superryanguo/pizza/pizza"
	"github.com/superryanguo/pizza/pizza/models"
	"github.com/superryanguo/pizza/pizza/services"
	"github.com/golang/glog"
)

type cartservice struct {
	cartRepo     models.CartRepository
	pizzaservice pizza.Service
}

func NewCartService(repo models.CartRepository, svc pizza.Service) services.CartService {

	return &cartservice{
		cartRepo:     repo,
		pizzaservice: svc,
	}
}

func (c cartservice) GetCart(ctx context.Context, userId string) (*[]models.CartQueryResult, error) {
	if userId == "" {
		glog.Error("Cannot get cart for empty user")
		return nil, errors.New("cannot get cart for empty user")
	}
	cart, err := c.cartRepo.GetCart(ctx, userId)
	if err != nil {
		glog.Errorf("Error getting cart %s", err)
		return nil, err
	}
	return cart, err
}

func (c cartservice) AddItem(ctx context.Context, itemId int, userId string, quantity int, price int) error {
	//See first if an item  is there in the users cart already if so return a response conflict
	item := c.cartRepo.GetCartItem(ctx, itemId, userId)
	if item != nil {
		glog.Errorf("The item already exists in the users cart")
		return errors.New("item-conflict")
	}
	//Get the price of the pizza and multiply the quantity
	pizza, err := c.pizzaservice.GetPizzaBYID(ctx, itemId)
	if err != nil {
		return err
	}
	itemPrice := int(pizza.Price) * quantity
	err = c.cartRepo.AddItem(ctx, itemId, userId, quantity, itemPrice)
	if err != nil {
		glog.Errorf("Error adding item to cart %s", err)
		return err
	}
	return nil
}

func (c cartservice) EditItem(ctx context.Context, cartItemId int, itemId int, quantity int, price int, userId string) error {
	pizza, err := c.pizzaservice.GetPizzaBYID(ctx, itemId)
	if err != nil {
		return err
	}
	itemPrice := int(pizza.Price) * quantity
	err = c.cartRepo.EditItem(ctx, cartItemId, itemId, quantity, itemPrice, userId)
	if err != nil {
		glog.Errorf("Error updating cart %s", err)
		return err
	}
	return nil
}

func (c cartservice) DeleteItem(ctx context.Context, cartItemId int, userId string) error {
	err := c.cartRepo.DeleteItem(ctx, cartItemId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (c cartservice) MakeItemInactive(ctx context.Context, cartItemID int) error {
	err := c.cartRepo.MakeItemInactive(ctx, cartItemID)
	return err
}
