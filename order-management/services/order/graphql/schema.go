package graphql

import (
	"context"
	"fmt"
	"time"

	"order-management/model"
	"order-management/services/order/service"

	gql "github.com/graphql-go/graphql"
)

type Server struct {
	svc    service.OrderService
	Schema gql.Schema
}

func NewServer(svc service.OrderService) (*Server, error) {
	s := &Server{svc: svc}
	schema, err := s.buildSchema()
	if err != nil {
		return nil, err
	}
	s.Schema = schema
	return s, nil
}

func (s *Server) buildSchema() (gql.Schema, error) {
	orderType := gql.NewObject(gql.ObjectConfig{
		Name: "Order",
		Fields: gql.Fields{
			"id":          &gql.Field{Type: gql.NewNonNull(gql.Int)},
			"customerId":  &gql.Field{Type: gql.NewNonNull(gql.Int)},
			"productId":   &gql.Field{Type: gql.NewNonNull(gql.Int)},
			"quantity":    &gql.Field{Type: gql.NewNonNull(gql.Int)},
			"unitPrice":   &gql.Field{Type: gql.NewNonNull(gql.Float)},
			"totalAmount": &gql.Field{Type: gql.NewNonNull(gql.Float)},
			"status":      &gql.Field{Type: gql.NewNonNull(gql.String)},
			"createdAt": &gql.Field{
				Type: gql.String,
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					if o, ok := p.Source.(*model.Order); ok && !o.CreatedAt.IsZero() {
						return o.CreatedAt.Format(time.RFC3339), nil
					}
					if o, ok := p.Source.(model.Order); ok && !o.CreatedAt.IsZero() {
						return o.CreatedAt.Format(time.RFC3339), nil
					}
					return time.Now().Format(time.RFC3339), nil
				},
			},
			"updatedAt": &gql.Field{
				Type: gql.String,
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					if o, ok := p.Source.(*model.Order); ok && !o.UpdatedAt.IsZero() {
						return o.UpdatedAt.Format(time.RFC3339), nil
					}
					if o, ok := p.Source.(model.Order); ok && !o.UpdatedAt.IsZero() {
						return o.UpdatedAt.Format(time.RFC3339), nil
					}
					return time.Now().Format(time.RFC3339), nil
				},
			},
		},
	})

	createOrderInput := gql.NewInputObject(gql.InputObjectConfig{
		Name: "CreateOrderInput",
		Fields: gql.InputObjectConfigFieldMap{
			"customerId": &gql.InputObjectFieldConfig{Type: gql.NewNonNull(gql.Int)},
			"productId":  &gql.InputObjectFieldConfig{Type: gql.NewNonNull(gql.Int)},
			"quantity":   &gql.InputObjectFieldConfig{Type: gql.NewNonNull(gql.Int)},
		},
	})

	updateStatusInput := gql.NewInputObject(gql.InputObjectConfig{
		Name: "UpdateOrderStatusInput",
		Fields: gql.InputObjectConfigFieldMap{
			"status": &gql.InputObjectFieldConfig{Type: gql.NewNonNull(gql.String)},
		},
	})

	query := gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"order": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id := int64(p.Args["id"].(int))
					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}
					return s.svc.GetOrder(ctx, id)
				},
			},
			"orders": &gql.Field{
				Type: gql.NewList(orderType),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}
					return s.svc.ListOrders(ctx)
				},
			},
		},
	})

	mutation := gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"createOrder": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"input": &gql.ArgumentConfig{Type: gql.NewNonNull(createOrderInput)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					req := model.CreateOrderRequest{
						CustomerID: int64(input["customerId"].(int)),
						ProductID:  int64(input["productId"].(int)),
						Quantity:   input["quantity"].(int),
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}
					return s.svc.CreateOrder(ctx, req)
				},
			},
			"updateOrderStatus": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"id":    &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
					"input": &gql.ArgumentConfig{Type: gql.NewNonNull(updateStatusInput)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id := int64(p.Args["id"].(int))
					input := p.Args["input"].(map[string]interface{})
					statusStr := input["status"].(string)

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					err := s.svc.UpdateOrderStatus(ctx, id, model.OrderStatus(statusStr))
					if err != nil {
						return nil, fmt.Errorf("failed to update status: %w", err)
					}
					return s.svc.GetOrder(ctx, id)
				},
			},
		},
	})

	return gql.NewSchema(gql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
}
