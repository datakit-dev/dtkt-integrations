package pkg

import (
	"bytes"
	"sort"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"

	htmlTemplate "html/template"
	textTemplate "text/template"
)

type CompiledTemplate struct {
	subject  *textTemplate.Template
	textBody *textTemplate.Template
	htmlBody *htmlTemplate.Template
}

type TemplateService struct {
	templates         map[string]*emailv1beta1.EmailTemplate
	compiledTemplates map[string]*CompiledTemplate
}

func NewTemplateService(templates []*emailv1beta1.EmailTemplate) (*TemplateService, error) {
	service := &TemplateService{
		templates:         make(map[string]*emailv1beta1.EmailTemplate),
		compiledTemplates: make(map[string]*CompiledTemplate),
	}

	for _, tmpl := range templates {
		if _, err := service.CreateTemplate(tmpl); err != nil {
			return nil, err
		}
	}

	return service, nil
}

func compileTemplate(template *emailv1beta1.EmailTemplate) (*CompiledTemplate, error) {
	if template.Language != emailv1beta1.TemplateLanguage_TEMPLATE_LANGUAGE_GO_TEMPLATE {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported template language: %s", template.Language)
	}

	subject, err := textTemplate.New("").Parse(template.Subject)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse subject template: %v", err)
	}

	textBody, err := textTemplate.New("").Parse(template.TextBody)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse text body template: %v", err)
	}

	htmlBody, err := htmlTemplate.New("").Parse(template.HtmlBody)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse HTML body template: %v", err)
	}

	return &CompiledTemplate{
		subject:  subject,
		textBody: textBody,
		htmlBody: htmlBody,
	}, nil
}

func (s *TemplateService) CreateTemplate(template *emailv1beta1.EmailTemplate) (*emailv1beta1.CreateEmailTemplateResponse, error) {
	if template.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "template ID cannot be empty")
	}

	if _, exists := s.templates[template.Id]; exists {
		return nil, status.Errorf(codes.AlreadyExists, "template with ID %s already exists", template.Id)
	}

	compiledTemplate, err := compileTemplate(template)
	if err != nil {
		return nil, err
	}

	s.templates[template.Id] = template
	s.compiledTemplates[template.Id] = compiledTemplate

	return &emailv1beta1.CreateEmailTemplateResponse{
		Template: template,
	}, nil
}

func (s *TemplateService) ListTemplates(pageSize int32, pageToken string) (*emailv1beta1.ListEmailTemplatesResponse, error) {
	keys := make([]string, 0, len(s.templates))
	for k := range s.templates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	start := 0
	if pageToken != "" {
		for i, id := range keys {
			if id == pageToken {
				start = i + 1 // Start after the last seen ID
				break
			}
		}
	}

	end := min(start+int(pageSize), len(keys))
	paged := keys[start:end]

	templates := make([]*emailv1beta1.EmailTemplate, 0, len(paged))
	for _, id := range paged {
		templates = append(templates, s.templates[id])
	}

	var nextPageToken string
	if end < len(keys) {
		nextPageToken = keys[end-1] // Last returned ID becomes the token
	}

	return &emailv1beta1.ListEmailTemplatesResponse{
		Templates:     templates,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *TemplateService) GetTemplate(id string) (*emailv1beta1.GetEmailTemplateResponse, error) {
	if template, exists := s.templates[id]; exists {
		return &emailv1beta1.GetEmailTemplateResponse{
			Template: template,
		}, nil
	}

	return nil, status.Errorf(codes.NotFound, "template with ID %s not found", id)
}

func (s *TemplateService) ExecuteTemplate(id string, data map[string]string) (string, string, string, error) {
	compiledTemplate, exists := s.compiledTemplates[id]
	if !exists {
		return "", "", "", status.Errorf(codes.NotFound, "template with ID %s not found", id)
	}

	subjectBuf := new(bytes.Buffer)
	textBodyBuf := new(bytes.Buffer)
	htmlBodyBuf := new(bytes.Buffer)

	err := compiledTemplate.subject.Execute(subjectBuf, data)
	if err != nil {
		return "", "", "", status.Errorf(codes.Internal, "failed to execute subject template: %v", err)
	}

	err = compiledTemplate.textBody.Execute(textBodyBuf, data)
	if err != nil {
		return "", "", "", status.Errorf(codes.Internal, "failed to execute text body template: %v", err)
	}

	err = compiledTemplate.htmlBody.Execute(htmlBodyBuf, data)
	if err != nil {
		return "", "", "", status.Errorf(codes.Internal, "failed to execute HTML body template: %v", err)
	}

	return subjectBuf.String(), textBodyBuf.String(), htmlBodyBuf.String(), nil
}

func (s *TemplateService) UpdateTemplate(template *emailv1beta1.EmailTemplate) (*emailv1beta1.UpdateEmailTemplateResponse, error) {
	if _, exists := s.templates[template.Id]; !exists {
		return nil, status.Errorf(codes.NotFound, "template with ID %s not found", template.Id)
	}

	compiledTemplate, err := compileTemplate(template)
	if err != nil {
		return nil, err
	}

	s.templates[template.Id] = template
	s.compiledTemplates[template.Id] = compiledTemplate

	return &emailv1beta1.UpdateEmailTemplateResponse{
		Template: template,
	}, nil
}

func (s *TemplateService) DeleteTemplate(id string) (*emailv1beta1.DeleteEmailTemplateResponse, error) {
	if _, exists := s.templates[id]; !exists {
		return nil, status.Errorf(codes.NotFound, "template with ID %s not found", id)
	}

	delete(s.templates, id)
	delete(s.compiledTemplates, id)

	return &emailv1beta1.DeleteEmailTemplateResponse{
		Id: id,
	}, nil
}
