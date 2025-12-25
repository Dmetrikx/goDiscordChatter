package bot

import (
	"testing"
)

func TestExtractProviderAndArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		defaultProvider string
		wantProvider    string
		wantArgs        []string
	}{
		{
			name:            "grok provider specified",
			args:            []string{"grok", "what", "is", "this"},
			defaultProvider: "openai",
			wantProvider:    "grok",
			wantArgs:        []string{"what", "is", "this"},
		},
		{
			name:            "openai provider specified",
			args:            []string{"openai", "tell", "me"},
			defaultProvider: "grok",
			wantProvider:    "openai",
			wantArgs:        []string{"tell", "me"},
		},
		{
			name:            "no provider specified",
			args:            []string{"what", "is", "this"},
			defaultProvider: "grok",
			wantProvider:    "grok",
			wantArgs:        []string{"what", "is", "this"},
		},
		{
			name:            "case insensitive provider",
			args:            []string{"GROK", "test"},
			defaultProvider: "openai",
			wantProvider:    "grok",
			wantArgs:        []string{"test"},
		},
		{
			name:            "empty args uses default",
			args:            []string{},
			defaultProvider: "openai",
			wantProvider:    "openai",
			wantArgs:        []string{},
		},
		{
			name:            "non-provider first arg",
			args:            []string{"something", "else"},
			defaultProvider: "grok",
			wantProvider:    "grok",
			wantArgs:        []string{"something", "else"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, gotArgs := extractProviderAndArgs(tt.args, tt.defaultProvider)

			if gotProvider != tt.wantProvider {
				t.Errorf("extractProviderAndArgs() provider = %v, want %v", gotProvider, tt.wantProvider)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("extractProviderAndArgs() args length = %v, want %v", len(gotArgs), len(tt.wantArgs))
				return
			}

			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("extractProviderAndArgs() args[%d] = %v, want %v", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestParseUserOpinionArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantProvider    string
		wantDays        int
		wantMaxMessages int
	}{
		{
			name:            "defaults with mention only",
			args:            []string{"<@123456>"},
			wantProvider:    "openai",
			wantDays:        DefaultUserOpinionDays,
			wantMaxMessages: DefaultUserOpinionMaxMessages,
		},
		{
			name:            "with provider",
			args:            []string{"<@123456>", "grok"},
			wantProvider:    "grok",
			wantDays:        DefaultUserOpinionDays,
			wantMaxMessages: DefaultUserOpinionMaxMessages,
		},
		{
			name:            "with provider and days",
			args:            []string{"<@123456>", "openai", "7"},
			wantProvider:    "openai",
			wantDays:        7,
			wantMaxMessages: DefaultUserOpinionMaxMessages,
		},
		{
			name:            "with all parameters",
			args:            []string{"<@123456>", "grok", "10", "500"},
			wantProvider:    "grok",
			wantDays:        10,
			wantMaxMessages: 500,
		},
		{
			name:            "only days specified",
			args:            []string{"<@123456>", "5"},
			wantProvider:    "openai",
			wantDays:        5,
			wantMaxMessages: DefaultUserOpinionMaxMessages,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, gotDays, gotMaxMessages := parseUserOpinionArgs(tt.args)

			if gotProvider != tt.wantProvider {
				t.Errorf("parseUserOpinionArgs() provider = %v, want %v", gotProvider, tt.wantProvider)
			}
			if gotDays != tt.wantDays {
				t.Errorf("parseUserOpinionArgs() days = %v, want %v", gotDays, tt.wantDays)
			}
			if gotMaxMessages != tt.wantMaxMessages {
				t.Errorf("parseUserOpinionArgs() maxMessages = %v, want %v", gotMaxMessages, tt.wantMaxMessages)
			}
		})
	}
}

func TestGetTopActiveUsers(t *testing.T) {
	tests := []struct {
		name       string
		userCounts map[string]int
		topN       int
		want       []string
	}{
		{
			name: "top 3 users",
			userCounts: map[string]int{
				"Alice":   10,
				"Bob":     5,
				"Charlie": 15,
				"Dave":    3,
				"Eve":     8,
			},
			topN: 3,
			want: []string{"Charlie", "Alice", "Eve"},
		},
		{
			name: "fewer users than topN",
			userCounts: map[string]int{
				"Alice": 10,
				"Bob":   5,
			},
			topN: 5,
			want: []string{"Alice", "Bob"},
		},
		{
			name:       "empty map",
			userCounts: map[string]int{},
			topN:       3,
			want:       []string{},
		},
		{
			name: "topN is 0",
			userCounts: map[string]int{
				"Alice": 10,
			},
			topN: 0,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTopActiveUsers(tt.userCounts, tt.topN)

			if len(got) != len(tt.want) {
				t.Errorf("getTopActiveUsers() length = %v, want %v", len(got), len(tt.want))
				return
			}

			// Create a map for easier comparison since order matters
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getTopActiveUsers()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestProviderDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{
			name:     "grok provider",
			provider: "grok",
			want:     "Grok",
		},
		{
			name:     "openai provider",
			provider: "openai",
			want:     "OpenAI",
		},
		{
			name:     "unknown provider",
			provider: "custom",
			want:     "Custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := providerDisplayName(tt.provider)
			if got != tt.want {
				t.Errorf("providerDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}
