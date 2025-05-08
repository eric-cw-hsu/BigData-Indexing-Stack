package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type Client struct {
	es    *elasticsearch.Client
	index string
}

func NewElasticSearchClient(address, username, password, index string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
		Username:  username,
		Password:  password,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{es: es, index: index}, nil
}

func (c *Client) InitIndex(mapping string) error {
	res, err := c.es.Indices.Exists([]string{c.index})
	if err != nil {
		return fmt.Errorf("index check failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		log.Printf("Index %s already exists", c.index)
		return nil
	}

	res, err = c.es.Indices.Create(c.index, c.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("index creation failed: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("index creation error: %s", res.String())
	}
	log.Printf("Index %s created successfully", c.index)

	return nil
}

func (c *Client) IndexDocument(id string, doc map[string]interface{}, routing string) {
	delete(doc, "_id")

	doc["join_field"] = map[string]interface{}{
		"name":   doc["fieldName"],
		"parent": doc["parentId"],
	}
	delete(doc, "parentId")
	delete(doc, "fieldName")

	body, _ := json.Marshal(doc)
	res, err := c.es.Index(
		c.index, bytes.NewReader(body),
		c.es.Index.WithDocumentID(id),
		c.es.Index.WithRouting(routing),
	)
	if err != nil {
		log.Printf("Failed to index document: %v", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error indexing document: %s", res.String())
		return
	}

	log.Printf("Document %s indexed successfully", id)
}

func (c *Client) DeleteDocument(id string) {
	res, err := c.es.Delete(c.index, id)
	if err != nil {
		log.Printf("Failed to delete document: %v", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error deleting document: %s", res.String())
		return
	}

	log.Printf("Document %s deleted successfully", id)
}
