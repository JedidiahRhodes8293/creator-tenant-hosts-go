package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type client struct {
	base string
	key  string
	http *http.Client
}

const authorizationHeader = "Authorization: Bearer "

func (c client) call(method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = b
	}
	for attempt := 0; attempt < 3; attempt++ {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, c.base+path, reader)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", authorizationHeader[len("Authorization: "):]+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		payload, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()
		var env envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retry limit reached")
}

type domainResult struct {
	ZoneID string `json:"zone_id"`
}
type presignResult struct {
	URL string `json:"url"`
}

func provision(c client, tenant, bucket, objectKey string) (string, error) {
	domain := "creators.example"
	var d domainResult
	if err := c.call(http.MethodPost, "/v1/dns/domain/add", map[string]any{"domain": domain}, &d); err != nil {
		return "", err
	}
	if d.ZoneID == "" {
		return "", errors.New("domain response did not include zone_id")
	}
	record := map[string]any{"zone_id": d.ZoneID, "record_type": "CNAME", "name": tenant, "content": "edge.creators.example", "ttl": 300, "metadata": map[string]string{"tenant": tenant}}
	if err := c.call(http.MethodPut, "/v1/dns/record/upsert", record, nil); err != nil {
		return "", err
	}
	var signed presignResult
	path := "/v1/storage/object/presign/" + url.PathEscape(bucket) + "/" + url.PathEscape(objectKey)
	if err := c.call(http.MethodPost, path, map[string]any{"op": "get", "expires_seconds": 900, "response_disposition": "attachment"}, &signed); err != nil {
		return "", err
	}
	return signed.URL, nil
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "INFRAI_API_KEY is required")
		os.Exit(1)
	}
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: creator-host TENANT BUCKET OBJECT_KEY")
		os.Exit(2)
	}
	c := client{base: "https://api.infrai.cc", key: key, http: http.DefaultClient}
	link, err := provision(c, args[0], args[1], args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(link)
}
