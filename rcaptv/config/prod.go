//go:build RELEASE
// +build RELEASE

package config

// All the code inside !IsProd if branches will be removed by the compiler in
// release builds.
const IsProd = true
