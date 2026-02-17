package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
)

func fromProtoProduct(p *v1.Product) *groupbuy.Product {
	if p == nil {
		return nil
	}

	var specs []*groupbuy.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &groupbuy.ProductSpec{
			ID:   s.Id,
			Name: s.Name,
		})
	}

	var rc *groupbuy.RoundingConfig
	if p.RoundingConfig != nil {
		rc = &groupbuy.RoundingConfig{
			Method: groupbuy.RoundingMethod(p.RoundingConfig.Method),
			Digit:  int(p.RoundingConfig.Digit),
		}
	}

	return &groupbuy.Product{
		ID:            p.Id,
		GroupBuyID:    p.GroupBuyId,
		Name:          p.Name,
		Description:   p.Description,
		ImageURL:      p.ImageUrl,
		PriceOriginal: p.PriceOriginal,
		ExchangeRate:  p.ExchangeRate,
		Rounding:      rc,
		PriceFinal:    p.PriceFinal,
		MaxQuantity:   p.MaxQuantity,
		Specs:         specs,
	}
}

func fromProtoShippingConfig(c *v1.ShippingConfig) *groupbuy.ShippingConfig {
	if c == nil {
		return nil
	}
	return &groupbuy.ShippingConfig{
		ID:    c.Id,
		Name:  c.Name,
		Type:  groupbuy.ShippingType(c.Type),
		Price: c.Price,
	}
}

func toProtoGroupBuy(gb *groupbuy.GroupBuy) *v1.GroupBuy {
	if gb == nil {
		return nil
	}

	var deadline *timestamppb.Timestamp
	if gb.Deadline != nil {
		deadline = timestamppb.New(*gb.Deadline)
	}

	return &v1.GroupBuy{
		Id:             gb.ID,
		Title:          gb.Title,
		Description:    gb.Description,
		CoverImageUrl:  gb.CoverImage,
		Status:         toProtoGroupBuyStatus(gb.Status),
		CreatedAt:      timestamppb.New(gb.CreatedAt),
		Deadline:       deadline,
		ExchangeRate:   gb.ExchangeRate,
		RoundingConfig: toProtoRoundingConfig(gb.Rounding),
		Creator:        toProtoUser(gb.Creator),

		SourceCurrency:  gb.SourceCurrency,
		ShippingConfigs: toProtoShippingConfigs(gb.ShippingConfigs),
	}
}

func toProtoRoundingConfig(r *groupbuy.RoundingConfig) *v1.RoundingConfig {
	if r == nil {
		return nil
	}
	return &v1.RoundingConfig{
		Method: toProtoRoundingMethod(r.Method),
		Digit:  safeIntToInt32(r.Digit),
	}
}

func toProtoShippingConfigs(configs []*groupbuy.ShippingConfig) []*v1.ShippingConfig {
	var res []*v1.ShippingConfig
	for _, c := range configs {
		res = append(res, toProtoShippingConfig(c))
	}
	return res
}

func toProtoShippingConfig(c *groupbuy.ShippingConfig) *v1.ShippingConfig {
	if c == nil {
		return nil
	}
	return &v1.ShippingConfig{
		Id:    c.ID,
		Name:  c.Name,
		Type:  toProtoShippingType(c.Type),
		Price: c.Price,
	}
}

func toProtoOrder(o *groupbuy.Order) *v1.Order {
	if o == nil {
		return nil
	}
	var items []*v1.OrderItem
	for _, i := range o.Items {
		items = append(items, &v1.OrderItem{
			Id:          i.ID,
			ProductId:   i.ProductID,
			SpecId:      i.SpecID,
			Quantity:    safeIntToInt32(i.Quantity),
			Status:      toProtoOrderItemStatus(i.Status),
			ProductName: i.ProductName,
			SpecName:    i.SpecName,
			Price:       i.Price,
		})
	}

	var pi *v1.PaymentInfo
	if o.PaymentInfo != nil {
		var paidAt *timestamppb.Timestamp
		if o.PaymentInfo.PaidAt != nil {
			paidAt = timestamppb.New(*o.PaymentInfo.PaidAt)
		}
		pi = &v1.PaymentInfo{
			Method:       o.PaymentInfo.Method,
			AccountLast5: o.PaymentInfo.AccountLast5,
			PaidAt:       paidAt,
			Amount:       o.PaymentInfo.Amount,
		}
	}

	return &v1.Order{
		Id:               o.ID,
		GroupBuyId:       o.GroupBuyID,
		UserId:           o.UserID,
		TotalAmount:      o.TotalAmount,
		PaymentStatus:    toProtoGroupBuyPaymentStatus(o.PaymentStatus),
		ContactInfo:      o.ContactInfo,
		ShippingAddress:  o.ShippingAddress,
		PaymentInfo:      pi,
		Items:            items,
		ShippingMethodId: o.ShippingMethodID,
		ShippingFee:      o.ShippingFee,
		Note:             o.Note,
	}
}

func toProtoProducts(products []*groupbuy.Product) []*v1.Product {
	var res []*v1.Product
	for _, p := range products {
		res = append(res, toProtoProduct(p))
	}
	return res
}

func toProtoProduct(p *groupbuy.Product) *v1.Product {
	if p == nil {
		return nil
	}
	var specs []*v1.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &v1.ProductSpec{Id: s.ID, Name: s.Name})
	}
	var rc *v1.RoundingConfig
	if p.Rounding != nil {
		rc = &v1.RoundingConfig{
			Method: toProtoRoundingMethod(p.Rounding.Method),
			Digit:  safeIntToInt32(p.Rounding.Digit),
		}
	}
	return &v1.Product{
		Id:             p.ID,
		GroupBuyId:     p.GroupBuyID,
		Name:           p.Name,
		Description:    p.Description,
		ImageUrl:       p.ImageURL,
		PriceOriginal:  p.PriceOriginal,
		ExchangeRate:   p.ExchangeRate,
		RoundingConfig: rc,
		PriceFinal:     p.PriceFinal,
		MaxQuantity:    p.MaxQuantity,
		Specs:          specs,
	}
}

func toProtoCategory(c *groupbuy.Category) *v1.Category {
	if c == nil {
		return nil
	}
	return &v1.Category{
		Id:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}

func toProtoPriceTemplate(pt *groupbuy.PriceTemplate) *v1.PriceTemplate {
	if pt == nil {
		return nil
	}
	var rc *v1.RoundingConfig
	if pt.Rounding != nil {
		rc = &v1.RoundingConfig{
			Method: toProtoRoundingMethod(pt.Rounding.Method),
			Digit:  safeIntToInt32(pt.Rounding.Digit),
		}
	}
	return &v1.PriceTemplate{
		Id:             pt.ID,
		Name:           pt.Name,
		SourceCurrency: pt.SourceCurrency,
		ExchangeRate:   pt.ExchangeRate,
		RoundingConfig: rc,
	}
}
