package console

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nwp/jenkins-cli/internal/client"
)

const pollInterval = time.Second

// Follow streams the full console output of a build, starting from the
// beginning, until the build completes. It writes to w.
func Follow(c *client.Client, jobURL, build string, w io.Writer) error {
	return stream(c, jobURL, build, 0, w)
}

// Tail streams only new console output from the build's current end position.
// It does NOT go back to the beginning; useful for watching in-progress builds.
func Tail(c *client.Client, jobURL, build string, w io.Writer) error {
	// Fetch current size to start from tail
	url := fmt.Sprintf("%s/%s/logText/progressiveText?start=0", jobURL, build)
	resp, err := c.GetRaw(url)
	if err != nil {
		return err
	}
	sizeStr := resp.Header.Get("X-Text-Size")
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	offset := int64(0)
	if sizeStr != "" {
		offset, _ = strconv.ParseInt(sizeStr, 10, 64)
	}
	return stream(c, jobURL, build, offset, w)
}

// GetText returns the full console text for a build (non-streaming).
func GetText(c *client.Client, jobURL, build string) (string, error) {
	return c.GetText(fmt.Sprintf("%s/%s/consoleText", jobURL, build))
}

func stream(c *client.Client, jobURL, build string, startOffset int64, w io.Writer) error {
	offset := startOffset
	for {
		url := fmt.Sprintf("%s/%s/logText/progressiveText?start=%d", jobURL, build, offset)
		resp, err := c.GetRaw(url)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read console: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("console HTTP %d", resp.StatusCode)
		}

		if len(body) > 0 {
			if _, err := w.Write(body); err != nil {
				return err
			}
		}

		sizeStr := resp.Header.Get("X-Text-Size")
		if sizeStr != "" {
			if n, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				offset = n
			}
		}

		moreData := resp.Header.Get("X-More-Data")
		if moreData != "true" {
			return nil
		}

		time.Sleep(pollInterval)
	}
}
