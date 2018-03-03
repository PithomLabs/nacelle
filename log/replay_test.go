package log

import (
	"github.com/aphistic/sweet"
	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

type ReplaySuite struct{}

func (s *ReplaySuite) TestReplay(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug)
	)

	adapter.LogWithFields(LevelDebug, map[string]interface{}{"x": "x"}, "foo", 12)
	adapter.LogWithFields(LevelDebug, map[string]interface{}{"y": "y"}, "bar", 43)
	adapter.LogWithFields(LevelDebug, map[string]interface{}{"z": "z"}, "baz", 74)
	adapter.Replay(LevelWarning)

	Expect(shim.messages).To(HaveLen(6))

	for i := 0; i < 3; i++ {
		Expect(shim.messages[i+0].level).To(Equal(LevelDebug))
		Expect(shim.messages[i+3].level).To(Equal(LevelWarning))
	}

	for i, format := range []string{"foo", "bar", "baz"} {
		Expect(shim.messages[i+0].format).To(Equal(format))
		Expect(shim.messages[i+3].format).To(Equal(format))
	}

	for i, expected := range []int{12, 43, 74} {
		Expect(shim.messages[i+0].args[0]).To(Equal(expected))
		Expect(shim.messages[i+3].args[0]).To(Equal(expected))
	}

	for i, field := range []string{"x", "y", "z"} {
		Expect(shim.messages[i+0].fields[field]).To(Equal(field))
		Expect(shim.messages[i+3].fields[field]).To(Equal(field))
	}
}

func (s *ReplaySuite) TestReplayTwice(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelDebug, "bar")
	adapter.Log(LevelDebug, "baz")
	adapter.Replay(LevelWarning)
	adapter.Replay(LevelError)

	Expect(shim.messages).To(HaveLen(9))
	Expect(shim.messages[0].level).To(Equal(LevelDebug))
	Expect(shim.messages[1].level).To(Equal(LevelDebug))
	Expect(shim.messages[2].level).To(Equal(LevelDebug))
	Expect(shim.messages[3].level).To(Equal(LevelWarning))
	Expect(shim.messages[4].level).To(Equal(LevelWarning))
	Expect(shim.messages[5].level).To(Equal(LevelWarning))
	Expect(shim.messages[6].level).To(Equal(LevelError))
	Expect(shim.messages[7].level).To(Equal(LevelError))
	Expect(shim.messages[8].level).To(Equal(LevelError))

	for i, format := range []string{"foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz"} {
		Expect(shim.messages[i].format).To(Equal(format))
	}
}

func (s *ReplaySuite) TestReplayAtHigherlevelNoops(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelDebug, "bar")
	adapter.Log(LevelDebug, "baz")
	adapter.Replay(LevelError)
	adapter.Replay(LevelWarning)

	Expect(shim.messages).To(HaveLen(6))
	Expect(shim.messages[0].level).To(Equal(LevelDebug))
	Expect(shim.messages[1].level).To(Equal(LevelDebug))
	Expect(shim.messages[2].level).To(Equal(LevelDebug))
	Expect(shim.messages[3].level).To(Equal(LevelError))
	Expect(shim.messages[4].level).To(Equal(LevelError))
	Expect(shim.messages[5].level).To(Equal(LevelError))

	for i, format := range []string{"foo", "bar", "baz", "foo", "bar", "baz"} {
		Expect(shim.messages[i].format).To(Equal(format))
	}
}

func (s *ReplaySuite) TestLogAfterReplaySendsImmediately(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelDebug, "bar")
	adapter.Log(LevelDebug, "baz")
	adapter.Replay(LevelWarning)
	adapter.Log(LevelDebug, "bnk")
	adapter.Log(LevelDebug, "qux")

	Expect(shim.messages).To(HaveLen(10))
	Expect(shim.messages[0].level).To(Equal(LevelDebug))
	Expect(shim.messages[1].level).To(Equal(LevelDebug))
	Expect(shim.messages[2].level).To(Equal(LevelDebug))
	Expect(shim.messages[3].level).To(Equal(LevelWarning))
	Expect(shim.messages[4].level).To(Equal(LevelWarning))
	Expect(shim.messages[5].level).To(Equal(LevelWarning))
	Expect(shim.messages[6].level).To(Equal(LevelDebug))
	Expect(shim.messages[7].level).To(Equal(LevelWarning))
	Expect(shim.messages[8].level).To(Equal(LevelDebug))
	Expect(shim.messages[9].level).To(Equal(LevelWarning))

	for i, format := range []string{"foo", "bar", "baz", "foo", "bar", "baz", "bnk", "bnk", "qux", "qux"} {
		Expect(shim.messages[i].format).To(Equal(format))
	}
}

func (s *ReplaySuite) TestLogAfterSecondReplaySendsAtNewLevel(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelDebug, "bar")
	adapter.Replay(LevelWarning)
	adapter.Replay(LevelError)
	adapter.Log(LevelDebug, "baz")
	adapter.Log(LevelDebug, "bnk")

	Expect(shim.messages).To(HaveLen(10))
	Expect(shim.messages[0].level).To(Equal(LevelDebug))
	Expect(shim.messages[1].level).To(Equal(LevelDebug))
	Expect(shim.messages[2].level).To(Equal(LevelWarning))
	Expect(shim.messages[3].level).To(Equal(LevelWarning))
	Expect(shim.messages[4].level).To(Equal(LevelError))
	Expect(shim.messages[5].level).To(Equal(LevelError))
	Expect(shim.messages[6].level).To(Equal(LevelDebug))
	Expect(shim.messages[7].level).To(Equal(LevelError))
	Expect(shim.messages[8].level).To(Equal(LevelDebug))
	Expect(shim.messages[9].level).To(Equal(LevelError))

	for i, format := range []string{"foo", "bar", "foo", "bar", "foo", "bar", "baz", "baz", "bnk", "bnk"} {
		Expect(shim.messages[i].format).To(Equal(format))
	}
}

func (s *ReplaySuite) TestCheckReplayAddsAttribute(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug, LevelInfo)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelInfo, "bar")
	adapter.Log(LevelDebug, "baz")
	adapter.Replay(LevelError)
	adapter.Log(LevelDebug, "bnk")

	Expect(shim.messages).To(HaveLen(8))
	Expect(shim.messages[0].fields).To(BeNil())
	Expect(shim.messages[1].fields).To(BeNil())
	Expect(shim.messages[2].fields).To(BeNil())
	Expect(shim.messages[3].fields[FieldReplay]).To(Equal(LevelDebug))
	Expect(shim.messages[4].fields[FieldReplay]).To(Equal(LevelInfo))
	Expect(shim.messages[5].fields[FieldReplay]).To(Equal(LevelDebug))
	Expect(shim.messages[6].fields).To(BeNil())
	Expect(shim.messages[7].fields[FieldReplay]).To(Equal(LevelDebug))
}

func (s *ReplaySuite) TestCheckSecondReplayAddsAttribute(t sweet.T) {
	var (
		shim    = &testShim{}
		clock   = glock.NewMockClock()
		adapter = newReplayAdapter(adaptShim(shim), clock, LevelDebug, LevelInfo)
	)

	adapter.Log(LevelDebug, "foo")
	adapter.Log(LevelInfo, "bar")
	adapter.Replay(LevelWarning)
	adapter.Replay(LevelError)
	adapter.Log(LevelDebug, "bnk")

	Expect(shim.messages).To(HaveLen(8))
	Expect(shim.messages[0].fields).To(BeNil())
	Expect(shim.messages[1].fields).To(BeNil())
	Expect(shim.messages[2].fields[FieldReplay]).To(Equal(LevelDebug))
	Expect(shim.messages[3].fields[FieldReplay]).To(Equal(LevelInfo))
	Expect(shim.messages[4].fields[FieldReplay]).To(Equal(LevelDebug))
	Expect(shim.messages[5].fields[FieldReplay]).To(Equal(LevelInfo))
	Expect(shim.messages[6].fields).To(BeNil())
	Expect(shim.messages[7].fields[FieldReplay]).To(Equal(LevelDebug))
}
