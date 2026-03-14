package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	aiv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/ai/v1beta1"
	"github.com/openai/openai-go/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var EmbeddingModels = []*aiv1beta1.EmbeddingModel{
	{
		Id:            openai.EmbeddingModelTextEmbeddingAda002,
		Name:          "Ada 002",
		MinDimensions: 1536,
		MaxDimensions: 1536,
	},
	{
		Id:            openai.EmbeddingModelTextEmbedding3Small,
		Name:          "Text Embedding 3 Small",
		MinDimensions: 256,
		MaxDimensions: 1536,
	},
	{
		Id:            openai.EmbeddingModelTextEmbedding3Large,
		Name:          "Text Embedding 3 Large",
		MinDimensions: 256,
		MaxDimensions: 3072,
	},
}

type EmbeddingService struct {
	aiv1beta1.EmbeddingServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewEmbeddingService(mux v1beta1.InstanceMux[*Instance]) *EmbeddingService {
	return &EmbeddingService{
		mux: mux,
	}
}

func (s *EmbeddingService) ListEmbeddingModels(context.Context, *aiv1beta1.ListEmbeddingModelsRequest) (*aiv1beta1.ListEmbeddingModelsResponse, error) {
	return &aiv1beta1.ListEmbeddingModelsResponse{
		Models: EmbeddingModels,
	}, nil
}

func (s *EmbeddingService) GetEmbeddingModel(_ context.Context, req *aiv1beta1.GetEmbeddingModelRequest) (*aiv1beta1.GetEmbeddingModelResponse, error) {
	for _, m := range EmbeddingModels {
		if m.Id == req.Id {
			return &aiv1beta1.GetEmbeddingModelResponse{
				Model: m,
			}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "embedding model not found")
}

func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, req *aiv1beta1.GenerateEmbeddingsRequest) (*aiv1beta1.GenerateEmbeddingsResponse, error) {
	model, err := s.GetEmbeddingModel(ctx, &aiv1beta1.GetEmbeddingModelRequest{
		Id: req.ModelId,
	})
	if err != nil {
		return nil, err
	}

	if req.Dimensions < model.Model.MinDimensions || req.Dimensions > model.Model.MaxDimensions {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dimensions: %d", req.Dimensions)
	}

	inputs := openai.EmbeddingNewParamsInputUnion{}
	for _, input := range req.Inputs {
		inputs.OfArrayOfStrings = append(inputs.OfArrayOfStrings, input.Text)
	}

	params := openai.EmbeddingNewParams{
		Model:          openai.EmbeddingModel(req.ModelId),
		Input:          inputs,
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	}

	if model.Model.MinDimensions != model.Model.MaxDimensions {
		params.Dimensions = openai.Int(int64(req.Dimensions))
	}

	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	res, err := inst.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}

	outputs := []*aiv1beta1.EmbeddingOutput{}
	for idx, data := range res.Data {
		outputs = append(outputs, &aiv1beta1.EmbeddingOutput{
			Input:   req.Inputs[idx],
			Vector:  data.Embedding,
			Success: true,
		})
	}

	return &aiv1beta1.GenerateEmbeddingsResponse{
		Outputs: outputs,
	}, nil
}
