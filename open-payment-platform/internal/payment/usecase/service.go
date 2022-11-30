package usecase

import (
	"context"

	"github.com/cloudwego/biz-demo/open-payment-platform/internal/payment/entity"
	"github.com/cloudwego/biz-demo/open-payment-platform/kitex_gen/payment"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewService)

var _ payment.PaymentSvc = (*Service)(nil)

type Service struct {
	repo Repository
}

func (s *Service) UnifyPay(ctx context.Context, req *payment.UnifyPayReq) (r *payment.UnifyPayResp, err error) {
	o := entity.NewOrder()
	o.PayWay = req.PayWay
	o.SubOpenid = req.SubOpenId
	o.MerchantID = req.MerchantId
	o.OutOrderNo = req.OutOrderNo
	o.Channel = "1"

	err = s.repo.Create(ctx, o)
	if err != nil {
		return nil, err
	}
	return &payment.UnifyPayResp{
		MerchantId:    o.MerchantID,
		SubMerchantId: o.SubOpenid,
		OutOrderNo:    o.OutOrderNo,
		JspayInfo:     "xxxxx",
		PayWay:        o.PayWay,
	}, nil
}

func (s *Service) uniqueOutOrderNo(ctx context.Context, outOrderNo string) (bool, error) {
	return true, nil
}

func (s *Service) QRPay(ctx context.Context, req *payment.QRPayReq) (r *payment.QRPayResp, err error) {
	onlyOne, err := s.uniqueOutOrderNo(ctx, req.OutOrderNo)
	if err != nil {
		klog.CtxErrorf(ctx, "err:%v", err)
		return nil, err
	}
	if !onlyOne {
		return nil, kerrors.NewBizStatusError(20000111, "订单号重复")
	}
	o := entity.NewOrder()
	o.PayWay = req.OutOrderNo
	o.TotalAmount = uint64(req.TotalAmount)
	o.MerchantID = req.MerchantId
	o.OutOrderNo = req.OutOrderNo
	o.Channel = "1"

	err = s.repo.Create(ctx, o)
	if err != nil {
		return nil, err
	}
	return &payment.QRPayResp{
		MerchantId:    o.MerchantID,
		SubMerchantId: o.SubOpenid,
		OutOrderNo:    o.OutOrderNo,
		PayWay:        o.PayWay,
	}, nil
}

func (s *Service) QueryOrder(ctx context.Context, req *payment.QueryOrderReq) (r *payment.QueryOrderResp, err error) {
	order, err := s.repo.GetByOutOrderNo(ctx, req.OutOrderNo)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &payment.QueryOrderResp{
		OrderStatus: order.OrderStatus,
	}, nil
}

func (s *Service) CloseOrder(ctx context.Context, req *payment.CloseOrderReq) (r *payment.CloseOrderResp, err error) {
	if err = s.repo.UpdateOrderStatus(ctx, req.OutOrderNo, 9); err != nil {
		return nil, err
	}
	return &payment.CloseOrderResp{}, nil
}

// NewService create new service
func NewService(r Repository) payment.PaymentSvc {
	return &Service{
		repo: r,
	}
}
