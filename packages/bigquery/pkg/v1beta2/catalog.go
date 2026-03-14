package v1beta2

import (
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta2 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta2"
)

type CatalogService struct {
	catalogv1beta2.UnimplementedCatalogServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewCatalogService(mux v1beta1.InstanceMux[*Instance]) *CatalogService {
	return &CatalogService{
		mux: mux,
	}
}

// func (s *CatalogService) CreateCatalog(ctx context.Context, req *catalogv1beta2.CreateCatalogRequest) (*catalogv1beta2.CreateCatalogResponse, error) {
// 	return nil, nil
// }

// func (s *CatalogService) DeleteCatalog(ctx context.Context, req *catalogv1beta2.DeleteCatalogRequest) (*catalogv1beta2.DeleteCatalogResponse, error) {
// 	return nil, nil
// }

// func (s *CatalogService) GetCatalog(ctx context.Context, req *catalogv1beta2.GetCatalogRequest) (*catalogv1beta2.GetCatalogResponse, error) {
// 	return nil, nil
// }

// func (s *CatalogService) ListCatalogs(ctx context.Context, req *catalogv1beta2.ListCatalogsRequest) (*catalogv1beta2.ListCatalogsResponse, error) {
// 	return nil, nil
// }

// func (s *CatalogService) UpdateCatalog(ctx context.Context, req *catalogv1beta2.UpdateCatalogRequest) (*catalogv1beta2.UpdateCatalogResponse, error) {
// 	return nil, nil
// }
