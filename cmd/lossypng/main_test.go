package main

import "testing"

func TestPathWithSuffix(t *testing.T) {
	for i, tt := range []struct {
		path   string
		suffix string
		want   string
	}{
		{"foo", "bar", "foobar"},
		{"foo.bar", "baz", "foobaz"},
		{"foo.bar.baz", "qux", "foo.barqux"},
	} {
		if got := pathWithSuffix(tt.path, tt.suffix); got != tt.want {
			t.Errorf("[%d] pathWithSuffix(%q, %q) = %q, want %q", i, tt.path, tt.suffix, got, tt.want)
		}
	}
}

func TestSizeDesc(t *testing.T) {
	for i, tt := range []struct {
		size int64
		want string
	}{
		{1, "1B"},
		{5, "5B"},
		{10, "10B"},
		{15, "15B"},
		{100, "100B"},
		{105, "105B"},
		{1000, "1000B"},
		{1005, "1005B"},
		{10000, "10kB"},
		{10005, "10kB"},
		{100000, "100kB"},
		{100005, "100kB"},
		{1000000, "1000kB"},
		{1000005, "1000kB"},
		{10000000, "10MB"},
		{10000005, "10MB"},
		{100000000, "100MB"},
		{100000005, "100MB"},
		{1000000000, "1000MB"},
		{1000000005, "1000MB"},
		{10000000000, "10GB"},
		{10000000005, "10GB"},
		{100000000000, "100GB"},
		{100000000005, "100GB"},
		{1000000000000, "1000GB"},
		{1000000000005, "1000GB"},
		{10000000000000, "10TB"},
		{10000000000005, "10TB"},
	} {
		if got := sizeDesc(tt.size); got != tt.want {
			t.Errorf("[%d] sizeDesc(%d) = %q, want %q", i, tt.size, got, tt.want)
		}
	}
}
