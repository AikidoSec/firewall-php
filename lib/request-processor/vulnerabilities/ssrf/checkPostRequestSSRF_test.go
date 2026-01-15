package ssrf

import (
	"main/context"
	"main/utils"
	"testing"
)

func TestCheckResolvedIpForSSRF_NoStoredInterceptorResult_ReturnsNil(t *testing.T) {
	inst := context.LoadForUnitTests(map[string]string{})
	context.ResetEventContext(inst)
	t.Cleanup(func() {
		context.ResetEventContext(inst)
		context.UnloadForUnitTests()
	})

	if res := CheckResolvedIpForSSRF(inst, "127.0.0.1"); res != nil {
		t.Fatalf("expected nil, got %#v", res)
	}
}

func TestCheckResolvedIpForSSRF_PublicIp_ReturnsNil(t *testing.T) {
	inst := context.LoadForUnitTests(map[string]string{})
	context.ResetEventContext(inst)
	t.Cleanup(func() {
		context.ResetEventContext(inst)
		context.UnloadForUnitTests()
	})

	ir := &utils.InterceptorResult{
		Operation: "curl_exec",
		Kind:      utils.Ssrf,
		Source:    "body",
		Metadata:  map[string]string{},
		Payload:   "http://example.test",
	}
	context.EventContextSetCurrentSsrfInterceptorResult(inst, ir)

	if res := CheckResolvedIpForSSRF(inst, "8.8.8.8"); res != nil {
		t.Fatalf("expected nil, got %#v", res)
	}
	if _, ok := ir.Metadata["isPrivateIp"]; ok {
		t.Fatalf("expected isPrivateIp to not be set for public IP")
	}
	if _, ok := ir.Metadata["resolvedIp"]; ok {
		t.Fatalf("expected resolvedIp to not be set for public IP")
	}
}

func TestCheckResolvedIpForSSRF_PrivateIp_ReturnsInterceptorResultWithMetadata(t *testing.T) {
	inst := context.LoadForUnitTests(map[string]string{})
	context.ResetEventContext(inst)
	t.Cleanup(func() {
		context.ResetEventContext(inst)
		context.UnloadForUnitTests()
	})

	ir := &utils.InterceptorResult{
		Operation: "curl_exec",
		Kind:      utils.Ssrf,
		Source:    "body",
		Metadata:  map[string]string{},
		Payload:   "http://example.test",
	}
	context.EventContextSetCurrentSsrfInterceptorResult(inst, ir)

	res := CheckResolvedIpForSSRF(inst, "127.0.0.1")
	if res == nil {
		t.Fatalf("expected non-nil interceptor result")
	}
	if res != ir {
		t.Fatalf("expected returned interceptor result to be the stored one")
	}
	if got := res.Metadata["resolvedIp"]; got != "127.0.0.1" {
		t.Fatalf("expected resolvedIp=127.0.0.1, got %q", got)
	}
	if got := res.Metadata["isPrivateIp"]; got != "true" {
		t.Fatalf("expected isPrivateIp=true, got %q", got)
	}
}

func TestCheckEffectiveHostnameForSSRF_PrivateIpHostname_ReturnsInterceptorResultWithMetadata(t *testing.T) {
	inst := context.LoadForUnitTests(map[string]string{})
	context.ResetEventContext(inst)
	t.Cleanup(func() {
		context.ResetEventContext(inst)
		context.UnloadForUnitTests()
	})

	ir := &utils.InterceptorResult{
		Operation: "curl_exec",
		Kind:      utils.Ssrf,
		Source:    "body",
		Metadata:  map[string]string{},
		Payload:   "http://example.test",
	}
	context.EventContextSetCurrentSsrfInterceptorResult(inst, ir)

	res := CheckEffectiveHostnameForSSRF(inst, "127.0.0.1")
	if res == nil {
		t.Fatalf("expected non-nil interceptor result")
	}
	if res != ir {
		t.Fatalf("expected returned interceptor result to be the stored one")
	}
	if got := res.Metadata["effectiveHostname"]; got != "127.0.0.1" {
		t.Fatalf("expected effectiveHostname=127.0.0.1, got %q", got)
	}
	if got := res.Metadata["resolvedIp"]; got != "127.0.0.1" {
		t.Fatalf("expected resolvedIp=127.0.0.1, got %q", got)
	}
	if got := res.Metadata["isPrivateIp"]; got != "true" {
		t.Fatalf("expected isPrivateIp=true, got %q", got)
	}
}

func TestCheckEffectiveHostnameForSSRF_IMDSHostname_ReturnsInterceptorResultWithIMDSMetadata(t *testing.T) {
	inst := context.LoadForUnitTests(map[string]string{})
	context.ResetEventContext(inst)
	t.Cleanup(func() {
		context.ResetEventContext(inst)
		context.UnloadForUnitTests()
	})

	ir := &utils.InterceptorResult{
		Operation: "curl_exec",
		Kind:      utils.Ssrf,
		Source:    "body",
		Metadata:  map[string]string{},
		Payload:   "http://example.test",
	}
	context.EventContextSetCurrentSsrfInterceptorResult(inst, ir)

	res := CheckEffectiveHostnameForSSRF(inst, "169.254.169.254")
	if res == nil {
		t.Fatalf("expected non-nil interceptor result")
	}
	if res != ir {
		t.Fatalf("expected returned interceptor result to be the stored one")
	}
	if got := res.Metadata["effectiveHostname"]; got != "169.254.169.254" {
		t.Fatalf("expected effectiveHostname=169.254.169.254, got %q", got)
	}
	if got := res.Metadata["resolvedIp"]; got != "169.254.169.254" {
		t.Fatalf("expected resolvedIp=169.254.169.254, got %q", got)
	}
	if got := res.Metadata["isIMDSIp"]; got != "true" {
		t.Fatalf("expected isIMDSIp=true, got %q", got)
	}
	// IMDS IPv4 is also in private ranges in our implementation.
	if got := res.Metadata["isPrivateIp"]; got != "true" {
		t.Fatalf("expected isPrivateIp=true, got %q", got)
	}
}
