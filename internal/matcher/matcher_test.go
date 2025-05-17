package matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Matcher unit tests using Ginkgo and Gomega
var _ = Describe("Matcher", func() {
	var m Matcher

	BeforeEach(func() {
		m = NewMatcher()
	})

	It("should match case-insensitive strings", func() {
		result := m.Match("TEst", "test")
		Expect(len(result)).To(Equal(1))
	})

	It("should match exact substring", func() {
		result := m.Match("foobar", "foo")
		Expect(len(result)).To(Equal(1))
	})

	It("should not match when keyword is not present", func() {
		result := m.Match("foobar", "baz")
		Expect(len(result)).To(Equal(0))
	})

	It("should match multiple occurrences", func() {
		result := m.Match("test test", "test")
		count := 0
		for i := 0; i < len("test test"); i++ {
			if result.IsMatch(i) {
				count++
			}
		}
		Expect(count).To(Equal(8)) // "test" appears twice, each 4 runes
	})

	It("should not match with empty keyword", func() {
		result := m.Match("foobar", "")
		Expect(len(result)).To(Equal(0))
	})
})
