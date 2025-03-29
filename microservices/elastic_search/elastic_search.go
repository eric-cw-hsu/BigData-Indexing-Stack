package main

import (
	"bytes"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

type ElasticSearchService struct {
	client *elasticsearch.Client
	logger *logrus.Logger
}

func NewElasticSearchService(logger *logrus.Logger) *ElasticSearchService {
	client, err := elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: []string{AppConfig.ElasticSearch.Addr},
			Username:  AppConfig.ElasticSearch.Username,
			Password:  AppConfig.ElasticSearch.Password,
		},
	)
	if err != nil {
		logger.Fatal("Failed to create ElasticSearch client: ", err)
	}

	return &ElasticSearchService{
		client: client,
		logger: logger,
	}
}

func (s *ElasticSearchService) IndexToElastic(index string, id string, data []byte) {
	res, err := s.client.Index(
		index,
		bytes.NewReader(data),
		s.client.Index.WithDocumentID(id),
	)
	if err != nil {
		s.logger.Error("Failed to index data to ElasticSearch: ", err)
		return
	}
	defer res.Body.Close()

	s.logger.Info("Data indexed in ElasticSearch")
}

func (s *ElasticSearchService) DeleteFromElastic(index, id string) {
	res, err := s.client.Delete(index, id)
	if err != nil {
		s.logger.Error("Failed to delete data from ElasticSearch: ", err)
		return
	}
	defer res.Body.Close()

	s.logger.Info("Data deleted from ElasticSearch")
}
