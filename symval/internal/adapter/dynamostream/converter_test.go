package dynamostream

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestConvertToDomainRecord(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		wantErr   bool
		errContains string
		validate  func(t *testing.T, record *events.DynamoDBEventRecord)
	}{
		{
			name: "full happy path - all fields present and valid",
			fixture: `{
				"eventID": "1",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-123" },
						"sk": { "S": "host.example.com" },
						"Owner": { "S": "alice@example.com" },
						"Type": { "S": "a" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result.GroupID != "grp-123" {
					t.Errorf("GroupID = %q, want %q", result.GroupID, "grp-123")
				}
				if result.Hostname != "host.example.com" {
					t.Errorf("Hostname = %q, want %q", result.Hostname, "host.example.com")
				}
				if result.Owner != "alice@example.com" {
					t.Errorf("Owner = %q, want %q", result.Owner, "alice@example.com")
				}
				if result.Type != symgroup.SymmetryType("a") {
					t.Errorf("Type = %q, want %q", result.Type, "a")
				}

				expectedTime, _ := time.Parse(time.RFC3339, "2025-10-30T12:34:56Z")
				if !result.ValidateTime.Equal(expectedTime) {
					t.Errorf("ValidateTime = %v, want %v", result.ValidateTime, expectedTime)
				}
			},
		},
		{
			name: "missing Owner field - should fail",
			fixture: `{
				"eventID": "2",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-456" },
						"sk": { "S": "host2.example.com" },
						"Type": { "S": "b" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: Owner",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for missing Owner, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "missing Type field - should fail",
			fixture: `{
				"eventID": "3",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-789" },
						"sk": { "S": "host3.example.com" },
						"Owner": { "S": "bob@example.com" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: Type",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for missing Type, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "missing ValidateTime field - should fail",
			fixture: `{
				"eventID": "4",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-abc" },
						"sk": { "S": "host4.example.com" },
						"Owner": { "S": "charlie@example.com" },
						"Type": { "S": "c" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: ValidateTime",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for missing ValidateTime, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "invalid RFC3339 time - should fail",
			fixture: `{
				"eventID": "5",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-def" },
						"sk": { "S": "host5.example.com" },
						"Owner": { "S": "dave@example.com" },
						"Type": { "S": "d" },
						"ValidateTime": { "S": "not-a-valid-time" }
					}
				}
			}`,
			wantErr: true,
			errContains: "invalid ValidateTime format",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for invalid time format, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "NewImage is nil (REMOVE event) - should error",
			fixture: `{
				"eventID": "6",
				"eventName": "REMOVE",
				"dynamodb": {
					"Keys": {
						"pk": { "S": "grp-ghi" },
						"sk": { "S": "host6.example.com" }
					}
				}
			}`,
			wantErr: true,
			errContains: "newImage is nil",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for nil NewImage, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "missing pk (GroupID) - should error",
			fixture: `{
				"eventID": "7",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"sk": { "S": "host7.example.com" },
						"Owner": { "S": "eve@example.com" },
						"Type": { "S": "e" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: GroupID",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for missing pk, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "missing sk (Hostname) - should error",
			fixture: `{
				"eventID": "8",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-jkl" },
						"Owner": { "S": "frank@example.com" },
						"Type": { "S": "f" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: Hostname",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for missing sk, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "empty pk value - should error",
			fixture: `{
				"eventID": "9",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "" },
						"sk": { "S": "host9.example.com" },
						"Owner": { "S": "grace@example.com" },
						"Type": { "S": "a" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: GroupID",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for empty pk, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
		{
			name: "empty sk value - should error",
			fixture: `{
				"eventID": "10",
				"eventName": "INSERT",
				"dynamodb": {
					"NewImage": {
						"pk": { "S": "grp-mno" },
						"sk": { "S": "" },
						"Owner": { "S": "henry@example.com" },
						"Type": { "S": "b" },
						"ValidateTime": { "S": "2025-10-30T12:34:56Z" }
					}
				}
			}`,
			wantErr: true,
			errContains: "missing required field: Hostname",
			validate: func(t *testing.T, record *events.DynamoDBEventRecord) {
				result, err := ConvertToDomainRecord(record.Change.NewImage)
				if err == nil {
					t.Fatal("expected error for empty sk, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result for error case, got %+v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSON fixture into DynamoDB event record
			var record events.DynamoDBEventRecord
			if err := json.Unmarshal([]byte(tt.fixture), &record); err != nil {
				t.Fatalf("failed to unmarshal fixture: %v", err)
			}

			// Run the validation function
			tt.validate(t, &record)
		})
	}
}

func TestConvertToDomainRecord_DirectNilInput(t *testing.T) {
	// Test calling the function directly with nil
	result, err := ConvertToDomainRecord(nil)
	if err == nil {
		t.Fatal("expected error for nil input, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for nil input, got %+v", result)
	}
}