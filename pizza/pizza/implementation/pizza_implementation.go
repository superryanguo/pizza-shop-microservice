package implementation

import (
	"context"

	"github.com/superryanguo/pizza/pizza"
	"github.com/superryanguo/pizza/pizza/models"
)

type pizzaservice struct {
	dbRepository models.PizzaRepository
}

func NewService(repo models.PizzaRepository) pizza.Service {
	return &pizzaservice{
		dbRepository: repo,
	}
}

func (s pizzaservice) GetPizzaBYID(ctx context.Context, id int) (pizza models.Pizza, err error) {
	p, err := s.dbRepository.GetPizzaByID(ctx, id)
	return p, err
}

func (s pizzaservice) GetAllPizzas(ctx context.Context, isVeg int) (pizza []models.Pizza, err error) {
	pizzas, err := s.dbRepository.GetAllPizzas(ctx, isVeg)
	return pizzas, err
}
