package pkg

import (
	"context"

	sheetsv1beta "github.com/datakit-dev/dtkt-integrations/sheets/pkg/proto/integration/sheets/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/protobuf/types/known/structpb"
)

type SpreadsheetService struct {
	sheetsv1beta.UnimplementedSpreadsheetServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewSpreadsheetService(mux v1beta1.InstanceMux[*Instance]) *SpreadsheetService {
	return &SpreadsheetService{
		mux: mux,
	}
}

func (s *SpreadsheetService) ListSpreadsheets(ctx context.Context, req *sheetsv1beta.ListSpreadsheetsRequest) (*sheetsv1beta.ListSpreadsheetsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	files, err := inst.drive.Files.List().Context(ctx).
		Q("mimeType='application/vnd.google-apps.spreadsheet'").
		Do()
	if err != nil {
		return nil, err
	}

	results := make([]*sheetsv1beta.Spreadsheet, len(files.Files))
	for idx, file := range files.Files {
		results[idx] = &sheetsv1beta.Spreadsheet{
			Id:   file.Id,
			Name: file.Name,
		}
	}

	return &sheetsv1beta.ListSpreadsheetsResponse{
		Spreadsheets: results,
	}, nil
}

func (s *SpreadsheetService) AppendSpreadsheetValues(ctx context.Context, req *sheetsv1beta.AppendSpreadsheetValuesRequest) (*sheetsv1beta.AppendSpreadsheetValuesResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	values := make([][]any, len(req.Values))
	for idx, row := range req.Values {
		values[idx] = row.AsSlice()
	}

	res, err := inst.sheets.Spreadsheets.Values.Append(
		req.SpreadsheetId,
		req.Range,
		&sheets.ValueRange{
			Values: values,
		},
	).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &sheetsv1beta.AppendSpreadsheetValuesResponse{
		UpdatedRows: res.Updates.UpdatedRows,
	}, nil
}

func (s *SpreadsheetService) GetSpreadsheetValues(ctx context.Context, req *sheetsv1beta.GetSpreadsheetValuesRequest) (*sheetsv1beta.GetSpreadsheetValuesResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	res, err := inst.sheets.Spreadsheets.Values.Get(req.SpreadsheetId, req.Range).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	values := make([]*structpb.ListValue, len(res.Values))
	for idx, value := range res.Values {
		list, err := structpb.NewList(value)
		if err != nil {
			return nil, err
		}

		values[idx] = list
	}

	return &sheetsv1beta.GetSpreadsheetValuesResponse{
		Values: values,
	}, nil
}
