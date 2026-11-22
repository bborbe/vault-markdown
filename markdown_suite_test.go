// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMarkdown(t *testing.T) {
	time.Local = time.UTC
	RegisterFailHandler(Fail)
	RunSpecs(t, "Markdown Suite")
}
