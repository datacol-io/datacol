package multiplexio

import (
	"bufio"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	firstTimeout = 100 * time.Millisecond
	timeout = 10 * time.Millisecond
}

const (
	line1          = "1 foo\n"
	line2          = "2 barbar\n"
	line3          = "3 quxquxqux\n"
	line4          = "4 bazbazbazbaz\n"
	unfinishedLine = "5 thisisnotacompletetoken"
)

func concatenatedStringsAsBytes(strings ...string) []byte {
	buf := make([]byte, 0, 1024)
	for _, s := range strings {
		buf = append(buf, []byte(s)...)
	}

	return buf
}

func readOneByteAtTheTime(src io.Reader, written *int) []byte {
	buf := make([]byte, 1024)
	for {
		n, err := io.ReadAtLeast(src, buf[*written:], 1)
		if err != nil {
			break
		}
		*written = *written + n
	}
	return buf[:*written]
}

// Ensure data is read from sources only when necessary
func TestLazyWrappedReaderFetching(t *testing.T) {
	var (
		pipeReader, pipeWriter = io.Pipe()
		reader                 = NewReader(
			Options{},
			Source{Reader: pipeReader},
		)
	)
	go func() {
		io.WriteString(pipeWriter, line1)
		io.WriteString(pipeWriter, line1)
		io.WriteString(pipeWriter, line1)
		pipeWriter.Close()
	}()
	// ask for enough bytes to get the first token
	io.CopyN(ioutil.Discard, reader, int64(len(line1)))
	go func() {
		// asking for one byte should consume only one more token
		io.CopyN(ioutil.Discard, reader, 1)
	}()
	// independently of our timeouts, give enough time to CopyN to block
	time.Sleep(10 * time.Millisecond)
	unconsumed, _ := io.Copy(ioutil.Discard, pipeReader)
	// leaving one token that we had no reason to fetch
	expected := int64(len(line1))
	if unconsumed != expected {
		t.Errorf("%v bytes unconsumed by the implementation, %v expected", unconsumed, expected)
	}
}

// Verify that a single reader so slow that it triggers all timeouts
// is forwarded correctly
func TestForwardingSingleVerySlowReader(t *testing.T) {
	var (
		pipeReader, pipeWriter = io.Pipe()
		reader                 = NewReader(
			Options{},
			Source{Reader: pipeReader},
		)
	)
	go func() {
		time.Sleep(2 * firstTimeout)
		io.WriteString(pipeWriter, line1)
		time.Sleep(2 * timeout)
		io.WriteString(pipeWriter, line2)
		time.Sleep(2 * timeout)
		pipeWriter.Close()
	}()
	written, _ := io.Copy(ioutil.Discard, reader)
	expected := int64(len(line1 + line2))
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
}

// Verify that one reader which is slow but within the read timeouts doesn't
// affect forwarding
func TestForwardingOneSlowReader(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
		time.Sleep(firstTimeout / 2)
		io.WriteString(pipeWriter2, line2)
		time.Sleep(timeout / 2)
		io.WriteString(pipeWriter2, line3)
		time.Sleep(timeout / 2)
		pipeWriter2.Close()
	}()
	written, _ := io.Copy(ioutil.Discard, reader)
	expected := int64(len(line1 + line2 + line3))
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
}

// Verify that one reader which is so slow that it triggers the timeouts
// doesn't affect forwarding
func TestForwardingOneVerySlowReader(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
		time.Sleep(2 * firstTimeout)
		io.WriteString(pipeWriter2, line2)
		time.Sleep(2 * timeout)
		io.WriteString(pipeWriter2, line3)
		time.Sleep(2 * timeout)
		pipeWriter2.Close()
	}()
	written, _ := io.Copy(ioutil.Discard, reader)
	expected := int64(len(line1 + line2 + line3))
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
}

// Verify that data from a reader which is not closed is forwarded correctly
// and that the aggregated reader blocks
func TestForwardingSingleHangingReader(t *testing.T) {
	var (
		pipeReader, pipeWriter = io.Pipe()
		reader                 = NewReader(
			Options{},
			Source{Reader: pipeReader},
		)
		written     int
		doneReading bool
	)
	go func() {
		io.WriteString(pipeWriter, line1)
		io.WriteString(pipeWriter, unfinishedLine)
	}()
	go func() {
		readOneByteAtTheTime(reader, &written)
		doneReading = true
	}()
	time.Sleep(2 * timeout)
	expected := len(line1)
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
	if doneReading {
		t.Errorf("reader expected to block but was done reading")
	}
}

// Verify that data from one reader which is not closed is forwarded correctly
// and that the aggregated reader blocks
func TestForwardingOneHangingReader(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		pipeReader3, _           = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
			Source{Reader: pipeReader3},
		)
		written     int
		doneReading bool
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
		io.WriteString(pipeWriter2, line2)
		pipeWriter2.Close()
	}()
	go func() {
		readOneByteAtTheTime(reader, &written)
		doneReading = true
	}()
	time.Sleep(2 * firstTimeout)
	expected := len(line1 + line2)
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
	if doneReading {
		t.Errorf("reader expected to block but was done reading")
	}
}

// Verify that a high volume, high frequency stream of tokens
// that should be ordered last does not take precedence over
// a low volume, low frequency stream of tokens that should
// be ordered first
func TestOrderingSequential(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
		expected = make([]byte, 0, 1024)
	)
	go func() {
		// exercise the initial waiting by delaying token
		// availability in the stream that should come first
		time.Sleep(firstTimeout / 2)
		for i := 0; i < 10; i++ {
			io.WriteString(pipeWriter1, line1)
			time.Sleep(timeout / 2)
		}
		pipeWriter1.Close()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			io.WriteString(pipeWriter2, line2)
		}
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	for i := 0; i < 10; i++ {
		expected = append(expected, []byte(line1)...)
	}
	for i := 0; i < 100; i++ {
		expected = append(expected, []byte(line2)...)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Ensure tokens from a low frequency stream can be interlaced with the ones
// from a high frequency stream
func TestOrderingInterlaced(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		// exercise the initial waiting by delaying token
		// availability in the stream that should come first
		time.Sleep(firstTimeout / 2)
		io.WriteString(pipeWriter1, line1)
		time.Sleep(timeout / 2)
		io.WriteString(pipeWriter1, line1)
		time.Sleep(timeout / 2)
		io.WriteString(pipeWriter1, line3)
		pipeWriter1.Close()
	}()
	go func() {
		io.WriteString(pipeWriter2, line2)
		io.WriteString(pipeWriter2, line4)
		io.WriteString(pipeWriter2, line4)
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line1,
		line1,
		line2,
		line3,
		line4,
		line4,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Verify that a stream for which the first token is slow to get but within
// the timeout is multiplexed correctly with a faster stream
func TestOrderingOneSlowStartReader(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter2, line2)
		pipeWriter2.Close()
		time.Sleep(firstTimeout / 2)
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line1,
		line2,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Verify that the aggregated stream does not wait for a stream which is so
// slow to extract a first token that it exceeds the timeout
func TestOrderingOneVerySlowStartReader(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter2, line2)
		pipeWriter2.Close()
		// if the reader is very slow, ordering
		// will not be guaranteed as we need to
		// move on
		time.Sleep(2 * firstTimeout)
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line2,
		line1,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Verify that a stream for which the tokens are slow to get but within the
// timeout is multiplexed correctly with a faster stream
func TestOrderingSlowReaders(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		time.Sleep(timeout / 2)
		io.WriteString(pipeWriter1, line2)
		pipeWriter1.Close()
	}()
	go func() {
		io.WriteString(pipeWriter2, line1)
		io.WriteString(pipeWriter2, line3)
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line1,
		line1,
		line2,
		line3,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Verify that the aggregated stream does not wait for a stream which is so
// slow to extract tokens that it exceeds the timeout
func TestOrderingVerySlowReaders(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		io.WriteString(pipeWriter2, line1)
		io.WriteString(pipeWriter2, line3)
		// if the reader is very slow, ordering
		// will not be guaranteed as we need to
		// move on
		time.Sleep(2 * timeout)
		io.WriteString(pipeWriter1, line2)
		pipeWriter1.Close()
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line1,
		line1,
		line3,
		line2,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Verify that the overall rate should not be limited by slow readers, once
// their timeout was exceeded once
func TestTimeoutsAreBySource(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe() // fast
		pipeReader2, pipeWriter2 = io.Pipe() // fast initial token, and then hangs
		pipeReader3, _           = io.Pipe() // hangs from the beginning
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
			Source{Reader: pipeReader3},
		)
		written int
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		io.WriteString(pipeWriter1, line2)
		io.WriteString(pipeWriter1, line3)
	}()
	go func() {
		io.WriteString(pipeWriter2, line1)
	}()
	go func() {
		readOneByteAtTheTime(reader, &written)
	}()
	blockingTime := firstTimeout + timeout
	time.Sleep(blockingTime + timeout/2) // add some margin
	expected := len(line1 + line1 + line2 + line3)
	if written != expected {
		t.Errorf("%v bytes copied, %v expected", written, expected)
	}
}

// Verify that a stream producing a token very late does not
// affect ordering of the well-behaving sources
func TestVerySlowReaderDoesNotPreemptOthers(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		pipeReader3, pipeWriter3 = io.Pipe()
		reader                   = NewReader(
			Options{},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
			Source{Reader: pipeReader3},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		time.Sleep(timeout / 2)
		io.WriteString(pipeWriter1, line2)
		pipeWriter1.Close()
	}()
	go func() {
		io.WriteString(pipeWriter2, line3)
		pipeWriter2.Close()
	}()
	go func() {
		time.Sleep(firstTimeout)
		io.WriteString(pipeWriter3, line4)
		pipeWriter3.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line1,
		line2,
		line3,
		line4,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Check that a custom Split function can be set
func TestCustomSplit(t *testing.T) {
	var (
		pipeReader, pipeWriter = io.Pipe()
		reader                 = NewReader(
			Options{Split: bufio.ScanWords},
			Source{Reader: pipeReader},
		)
	)
	go func() {
		io.WriteString(pipeWriter, "1 2 3")
		pipeWriter.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := []byte("1\n2\n3\n")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Check that a custom Less function can be set
func TestCustomLess(t *testing.T) {
	var (
		ByDescLenLess            = func(i, j []byte) bool { return len(i) > len(j) }
		pipeReader1, pipeWriter1 = io.Pipe()
		pipeReader2, pipeWriter2 = io.Pipe()
		reader                   = NewReader(
			Options{Less: ByDescLenLess},
			Source{Reader: pipeReader1},
			Source{Reader: pipeReader2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line3)
		io.WriteString(pipeWriter2, line4)
		io.WriteString(pipeWriter2, line2)
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	expected := concatenatedStringsAsBytes(
		line4,
		line3,
		line2,
		line1,
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}

// Check that a custom Write function can be set
func TestCustomWrites(t *testing.T) {
	var (
		pipeReader1, pipeWriter1 = io.Pipe()
		Write1                   = func(dest io.Writer, token []byte) (n int, err error) {
			return io.WriteString(dest, "ZZZZ<"+string(token)+">\n")
		}
		pipeReader2, pipeWriter2 = io.Pipe()
		Write2                   = func(dest io.Writer, token []byte) (n int, err error) {
			return io.WriteString(dest, "AAAA<"+string(token)+">\n")
		}
		reader = NewReader(
			Options{},
			Source{Reader: pipeReader1, Write: Write1},
			Source{Reader: pipeReader2, Write: Write2},
		)
	)
	go func() {
		io.WriteString(pipeWriter1, line1)
		pipeWriter1.Close()
		io.WriteString(pipeWriter2, line2)
		pipeWriter2.Close()
	}()
	actual := readOneByteAtTheTime(reader, new(int))
	// Write2's prefix is "before" Write1's but it must not affect
	// affect ordering
	expected := concatenatedStringsAsBytes(
		"ZZZZ<",
		strings.Trim(line1, "\n"),
		">\n",
		"AAAA<",
		strings.Trim(line2, "\n"),
		">\n",
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("`%v` read, `%v` expected", string(actual), string(expected))
	}
}
