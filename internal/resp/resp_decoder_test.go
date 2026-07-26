package resp

import (
	"reflect"
	"testing"
)

// I am already stripping 1st character like :,+,$ in my implementation,
// so the input does look a bit incomplete but its correct as per our current
//implementation

func assertRESPValue(t *testing.T, got, want *RESPValue) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\nexpected: %#v\n     got: %#v", want, got)
	}
}

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  *RESPValue
		done    bool
		wantErr bool
	}{
		{
			name:  "positive",
			input: "123\r\n",
			output: &RESPValue{
				Type:  RESPInteger,
				Value: 123,
			},
			done: true,
		},
		{
			name:  "negative",
			input: "-99\r\n",
			output: &RESPValue{
				Type:  RESPInteger,
				Value: -99,
			},
			done: true,
		},
		{
			name:   "incomplete",
			input:  "123\r",
			done:   false,
			output: nil,
		},
		{
			name:    "invalid",
			input:   "12abc\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := decodeInt([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}

func TestDecodeSimpleString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output *RESPValue
		done   bool
	}{
		{
			name:  "normal",
			input: "OK\r\n",
			output: &RESPValue{
				Type:  RESPSimpleString,
				Value: "OK",
			},
			done: true,
		},
		{
			name:  "empty",
			input: "\r\n",
			output: &RESPValue{
				Type:  RESPSimpleString,
				Value: "",
			},
			done: true,
		},
		{
			name:   "incomplete",
			input:  "OK\r",
			done:   false,
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := decodeSimpleString([]byte(tt.input))

			if err != nil {
				t.Fatal(err)
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}

func TestDecodeError(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output *RESPValue
		done   bool
	}{
		{
			name:  "normal",
			input: "ERR unknown command\r\n",
			output: &RESPValue{
				Type:  RESPError,
				Value: "ERR unknown command",
			},
			done: true,
		},
		{
			name:   "incomplete",
			input:  "ERR unknown command\r",
			done:   false,
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := decodeError([]byte(tt.input))

			if err != nil {
				t.Fatal(err)
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}

func TestDecodeBulkString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  *RESPValue
		done    bool
		wantErr bool
	}{
		{
			name:  "normal",
			input: "5\r\nhello\r\n",
			output: &RESPValue{
				Type:  RESPBulkString,
				Value: "hello",
			},
			done: true,
		},
		{
			name:  "empty",
			input: "0\r\n\r\n",
			output: &RESPValue{
				Type:  RESPBulkString,
				Value: "",
			},
			done: true,
		},
		{
			name:   "null",
			input:  "-1\r\n",
			output: nil,
			done:   true,
		},
		{
			name:   "incomplete",
			input:  "5\r\nhel",
			output: nil,
			done:   false,
		},
		{
			name:    "negative length",
			input:   "-5\r\n",
			wantErr: true,
		},
		{
			name:    "missing crlf",
			input:   "5\r\nhelloXX",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := decodeBulkString([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}

func TestDecodeArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  *RESPValue
		done    bool
		wantErr bool
	}{
		{
			name:  "empty array",
			input: "0\r\n",
			output: &RESPValue{
				Type:  RESPArray,
				Value: []*RESPValue{},
			},
			done: true,
		},
		{
			name:   "null array",
			input:  "-1\r\n",
			output: nil,
			done:   true,
		},
		{
			name:  "bulk strings",
			input: "2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n",
			output: &RESPValue{
				Type: RESPArray,
				Value: []*RESPValue{
					{
						Type:  RESPBulkString,
						Value: "PING",
					},
					{
						Type:  RESPBulkString,
						Value: "PONG",
					},
				},
			},
			done: true,
		},
		{
			name:  "nested array",
			input: "2\r\n:1\r\n*1\r\n+OK\r\n",
			output: &RESPValue{
				Type: RESPArray,
				Value: []*RESPValue{
					{
						Type:  RESPInteger,
						Value: 1,
					},
					{
						Type: RESPArray,
						Value: []*RESPValue{
							{
								Type:  RESPSimpleString,
								Value: "OK",
							},
						},
					},
				},
			},
			done: true,
		},
		{
			name:    "negative length",
			input:   "-5\r\n",
			wantErr: true,
		},
		{
			name:   "incomplete",
			input:  "2\r\n$4\r\nPING\r\n",
			done:   false,
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := decodeArray([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  *RESPValue
		done    bool
		wantErr bool
	}{
		{
			name:  "simple string",
			input: "+OK\r\n",
			output: &RESPValue{
				Type:  RESPSimpleString,
				Value: "OK",
			},
			done: true,
		},
		{
			name:  "integer",
			input: ":42\r\n",
			output: &RESPValue{
				Type:  RESPInteger,
				Value: 42,
			},
			done: true,
		},
		{
			name:  "bulk string",
			input: "$5\r\nhello\r\n",
			output: &RESPValue{
				Type:  RESPBulkString,
				Value: "hello",
			},
			done: true,
		},
		{
			name:   "null bulk string",
			input:  "$-1\r\n",
			output: nil,
			done:   true,
		},
		{
			name:  "array",
			input: "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n",
			output: &RESPValue{
				Type: RESPArray,
				Value: []*RESPValue{
					{
						Type:  RESPBulkString,
						Value: "PING",
					},
					{
						Type:  RESPBulkString,
						Value: "PONG",
					},
				},
			},
			done: true,
		},
		{
			name:  "nested array",
			input: "*2\r\n:1\r\n*1\r\n+OK\r\n",
			output: &RESPValue{
				Type: RESPArray,
				Value: []*RESPValue{
					{
						Type:  RESPInteger,
						Value: 1,
					},
					{
						Type: RESPArray,
						Value: []*RESPValue{
							{
								Type:  RESPSimpleString,
								Value: "OK",
							},
						},
					},
				},
			},
			done: true,
		},
		{
			name:  "error",
			input: "-ERR unknown command\r\n",
			output: &RESPValue{
				Type:  RESPError,
				Value: "ERR unknown command",
			},
			done: true,
		},
		{
			name:    "unsupported type",
			input:   "?abc",
			wantErr: true,
		},
		{
			name:   "empty input",
			input:  "",
			output: nil,
			done:   false,
		},
		{
			name:   "incomplete bulk string",
			input:  "$5\r\nhel",
			output: nil,
			done:   false,
		},
		{
			name:   "incomplete array",
			input:  "*2\r\n$4\r\nPING\r\n",
			output: nil,
			done:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, done, err := Decode([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			assertRESPValue(t, got, tt.output)
		})
	}
}