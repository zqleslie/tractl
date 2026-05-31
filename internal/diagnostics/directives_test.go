package diagnostics

import "testing"

func TestValidKinds_AllSixPresent(t *testing.T) {
	for _, k := range []string{KindTLS, KindTCP, KindTransport, KindLifecycle, KindDNS, KindConnection} {
		if _, ok := ValidKinds[k]; !ok {
			t.Errorf("ValidKinds missing %q", k)
		}
	}
	if len(ValidKinds) != 6 {
		t.Errorf("expected len(ValidKinds) == 6, got %d", len(ValidKinds))
	}
}

func TestValidKinds_UnknownStringAbsent(t *testing.T) {
	for _, k := range []string{"wireshark", "nmap", ""} {
		if _, ok := ValidKinds[k]; ok {
			t.Errorf("ValidKinds should not contain %q", k)
		}
	}
}

func TestValidRetentions_AllThreePresent(t *testing.T) {
	for _, r := range []string{RetentionStep, RetentionWorkflow, RetentionSpec} {
		if _, ok := ValidRetentions[r]; !ok {
			t.Errorf("ValidRetentions missing %q", r)
		}
	}
	if len(ValidRetentions) != 3 {
		t.Errorf("expected len(ValidRetentions) == 3, got %d", len(ValidRetentions))
	}
}

func TestValidRetentions_UnknownStringAbsent(t *testing.T) {
	for _, r := range []string{"session", "global", ""} {
		if _, ok := ValidRetentions[r]; ok {
			t.Errorf("ValidRetentions should not contain %q", r)
		}
	}
}

func TestStepDiagnosticsOutput_ZeroValueIsValid(t *testing.T) {
	var out StepDiagnosticsOutput
	if out.Events != nil {
		t.Errorf("expected Events to be nil in zero value, got %v", out.Events)
	}
	if out.EnabledKinds != nil {
		t.Errorf("expected EnabledKinds to be nil in zero value, got %v", out.EnabledKinds)
	}
}

func TestDiagnosticsEvent_DetailNotNilByConvention(t *testing.T) {
	ev := DiagnosticsEvent{
		MonotonicMs: 42,
		Kind:        KindTLS,
		Detail:      map[string]string{},
	}
	if ev.Detail == nil {
		t.Error("Detail must be non-nil per producer convention")
	}
}

func TestKindConstants_MirrorSpecValues(t *testing.T) {
	cases := []struct{ got, want string }{
		{KindTLS, "tls"},
		{KindTCP, "tcp"},
		{KindTransport, "transport"},
		{KindLifecycle, "lifecycle"},
		{KindDNS, "dns"},
		{KindConnection, "connection"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("constant value mismatch: got %q, want %q", c.got, c.want)
		}
	}
}

func TestRetentionConstants_MirrorSpecValues(t *testing.T) {
	cases := []struct{ got, want string }{
		{RetentionStep, "step"},
		{RetentionWorkflow, "workflow"},
		{RetentionSpec, "spec"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("constant value mismatch: got %q, want %q", c.got, c.want)
		}
	}
}
