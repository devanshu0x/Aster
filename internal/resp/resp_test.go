package resp

import (
	"reflect"
	"testing"
)

// I am already stripping 1st character like :,+,$ in my implementation,
// so the input does look a bit incomplete but its correct as per our current
//implementation

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  int
		done    bool
		wantErr bool
	}{
		{
			name:    "positive number",
			input:   "124\r\n",
			output:  124,
			done:    true,
			wantErr: false,
		},
		{
			name:    "negative number",
			input:   "-12378\r\n",
			output:  -12378,
			done:    true,
			wantErr: false,
		},
		{
			name:    "incomplete RESP number",
			input:   "1234\r",
			output:  0,
			done:    false,
			wantErr: false,
		},
		{
			name:    "invalid integer",
			input:   "12abc\r\n",
			output:  0,
			done:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _, done, err := decodeInt([]byte(tt.input))

			// Check whether an error was expected.
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}

			// If an error occurred (and was expected), we're done.
			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v, got %v", tt.done, done)
			}

			if v != tt.output {
				t.Fatalf("expected value=%d, got %d", tt.output, v)
			}
		})
	}
}

func TestDecodeSimpleString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  string
		done    bool
		wantErr bool
	}{
		{
			name:    "normal string",
			input:   "OK\r\n",
			output:  "OK",
			done:    true,
		},
		{
			name:    "empty string",
			input:   "\r\n",
			output:  "",
			done:    true,
		},
		{
			name:    "incomplete",
			input:   "OK\r",
			output:  "",
			done:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _, done, err := decodeSimpleString([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			if v != tt.output {
				t.Fatalf("expected %q got %q", tt.output, v)
			}
		})
	}
}

func TestDecodeError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  string
		done    bool
		wantErr bool
	}{
		{
			name:    "normal error",
			input:   "ERR unknown command\r\n",
			output:  "ERR unknown command",
			done:    true,
		},
		{
			name:    "incomplete",
			input:   "ERR unknown command\r",
			output:  "",
			done:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _, done, err := decodeError([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			if v != tt.output {
				t.Fatalf("expected %q got %q", tt.output, v)
			}
		})
	}
}

func TestDecodeBulkString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  Value
		done    bool
		wantErr bool
	}{
		{
			name:    "normal",
			input:   "5\r\nhello\r\n",
			output:  "hello",
			done:    true,
		},
		{
			name:    "empty",
			input:   "0\r\n\r\n",
			output:  "",
			done:    true,
		},
		{
			name:    "null bulk string",
			input:   "-1\r\n",
			output:  nil,
			done:    true,
		},
		{
			name:    "incomplete",
			input:   "5\r\nhel",
			output:  "",
			done:    false,
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
			v, _, done, err := decodeBulkString([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			if !reflect.DeepEqual(v, tt.output) {
				t.Fatalf("expected %#v got %#v", tt.output, v)
			}
		})
	}
}


func TestDecodeArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  Value
		done    bool
		wantErr bool
	}{
		{
			name:   "empty array",
			input:  "0\r\n",
			output: []Value{},
			done:   true,
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
			output: []Value{
				"PING",
				"PONG",
			},
			done: true,
		},
		{
			name:  "nested array",
			input: "2\r\n:1\r\n*1\r\n+OK\r\n",
			output: []Value{
				1,
				[]Value{"OK"},
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
			output: nil,
			done:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _, done, err := decodeArray([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			if !reflect.DeepEqual(v, tt.output) {
				t.Fatalf("expected %#v got %#v", tt.output, v)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  Value
		done    bool
		wantErr bool
	}{
		{
			name:    "simple string",
			input:   "+OK\r\n",
			output:  "OK",
			done:    true,
		},
		{
			name:    "integer",
			input:   ":42\r\n",
			output:  42,
			done:    true,
		},
		{
			name:    "bulk string",
			input:   "$5\r\nhello\r\n",
			output:  "hello",
			done:    true,
		},
		{
			name:    "null bulk string",
			input:   "$-1\r\n",
			output:  nil,
			done:    true,
		},
		{
			name:   "array",
			input:  "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n",
			output: []Value{"PING", "PONG"},
			done:   true,
		},
		{
			name:   "nested array",
			input:  "*2\r\n:1\r\n*1\r\n+OK\r\n",
			output: []Value{1, []Value{"OK"}},
			done:   true,
		},
		{
			name:    "error type",
			input:   "-ERR unknown command\r\n",
			output:  "ERR unknown command",
			done:    true,
		},
		{
			name:    "unsupported type",
			input:   "?abc",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "incomplete bulk string",
			input:   "$5\r\nhel",
			output:  "",
			done:    false,
		},
		{
			name:    "incomplete array",
			input:   "*2\r\n$4\r\nPING\r\n",
			done:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _, done, err := Decode([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v got=%v", tt.wantErr, err)
			}

			if err != nil {
				return
			}

			if done != tt.done {
				t.Fatalf("expected done=%v got=%v", tt.done, done)
			}

			if !reflect.DeepEqual(v, tt.output) {
				t.Fatalf("expected %#v got %#v", tt.output, v)
			}
		})
	}
}