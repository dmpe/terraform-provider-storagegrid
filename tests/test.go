// Copyright (c) HashiCorp, Inc.

package main

import (
    "encoding/json"
    "fmt"
)

// PolicyStatement struct with custom Resource type to handle both string and []string
type PolicyStatement struct {
    Effect   string    `json:"Effect"`
    Action   string    `json:"Action"`
    Resource Resource  `json:"Resource"` // Custom type to handle string or []string
}

// Resource type to handle both string and []string
type Resource struct {
    Values []string
}

// Custom UnmarshalJSON function for Resource
func (r *Resource) UnmarshalJSON(data []byte) error {
    // Try to unmarshal into a single string
    var singleValue string
    if err := json.Unmarshal(data, &singleValue); err == nil {
        r.Values = []string{singleValue}
        return nil
    }

    // If that fails, try to unmarshal into a slice of strings
    var arrayValue []string
    if err := json.Unmarshal(data, &arrayValue); err == nil {
        r.Values = arrayValue
        return nil
    }

    return fmt.Errorf("failed to unmarshal Resource field")
}

// S3Policy struct containing PolicyStatement
type S3Policy struct {
    Statement []PolicyStatement `json:"Statement"`
}

// Group struct including S3Policy
type Group struct {
    ID        string   `json:"id"`
    AccountID string   `json:"accountId"`
    Policies  S3Policy `json:"s3"`
}

func main() {
    // JSON data with Resource as a string
    jsonDataSingle := `{
        "id": "aec838fe-523f-bd43-a4df-33ebde931352",
        "accountId": "10554250556339739676",
        "s3": {
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "s3:*",
                    "Resource": "arn:aws:s3:::*"
                }
            ]
        }
    }`

    // JSON data with Resource as an array
    jsonDataArray := `{
        "id": "aec838fe-523f-bd43-a4df-33ebde931352",
        "accountId": "10554250556339739676",
        "s3": {
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "s3:*",
                    "Resource": [
                        "arn:aws:s3:::bucket-test",
                        "arn:aws:s3:::bucket-test/*"
                    ]
                }
            ]
        }
    }`

    var group Group

    // Parsing single string Resource JSON
    err := json.Unmarshal([]byte(jsonDataSingle), &group)
    if err != nil {
        fmt.Println("Error parsing single Resource JSON:", err)
    } else {
        fmt.Printf("Parsed single Resource JSON: %+v\n", group)
    }

    // Parsing array Resource JSON
    err = json.Unmarshal([]byte(jsonDataArray), &group)
    if err != nil {
        fmt.Println("Error parsing array Resource JSON:", err)
    } else {
        fmt.Printf("Parsed array Resource JSON: %+v\n", group)
    }
}
