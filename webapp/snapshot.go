package webapp

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
)

func snapshotWrapper(c *gin.Context) {
	sid, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	bid, ok := c.GetQuery("bid")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", bid).First(&b).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.BookmarkID != b.ID {
		setNotification(c, nError, "Invalid bookmark ID", false)
		return
	}
	err = model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.Size == 0 {
		s.Size = storage.GetSnapshotSize(s.Key)
		err = model.DB.Save(s).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			return
		}
	}
	var otherSnapshots []struct {
		Title string
		Bid   int64
		Sid   string
	}
	err = model.DB.
		Model(&model.Snapshot{}).
		Select("bookmarks.id as bid, snapshots.key as sid, snapshots.title as title").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("bookmarks.url = ? and snapshots.key != ?", b.URL, s.Key).Find(&otherSnapshots).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	render(c, http.StatusOK, "snapshotWrapper", map[string]interface{}{
		"Bookmark":       b,
		"Snapshot":       s,
		"hideFooter":     true,
		"OtherSnapshots": otherSnapshots,
	})
}

func downloadSnapshot(c *gin.Context) {
	id, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	r, err := storage.GetSnapshot(id)
	if err != nil {
		return
	}
	defer r.Close()
	gr, err := gzip.NewReader(r)
	if err != nil {
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=omnom_snapshot_%s.html;", id))
	c.Status(http.StatusOK)
	doc := html.NewTokenizer(gr)
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if err == io.EOF {
				return
			}
			// TODO error handling
			return
		case html.SelfClosingTagToken:
		case html.StartTagToken:
			c.Writer.Write([]byte("<"))
			tn, hasAttr := doc.TagName()
			c.Writer.Write(tn)
			if hasAttr {
				generateTagAttributes(tn, doc, c.Writer)
			}
			c.Writer.Write([]byte(">"))
		case html.TextToken:
			c.Writer.Write(doc.Text())
		case html.EndTagToken:
			tn, _ := doc.TagName()
			c.Writer.Write([]byte(fmt.Sprintf(`</%s>`, tn)))
		}
	}
}

func generateTagAttributes(tagName []byte, doc *html.Tokenizer, out io.Writer) {
	for {
		aName, aVal, moreAttr := doc.TagAttr()
		out.Write([]byte(fmt.Sprintf(` %s="`, aName)))
		if attributeHasResource(aName, aVal) {
			href := string(aVal)
			res, err := storage.GetResource(filepath.Base(href))
			if err == nil {
				gres, err := gzip.NewReader(res)
				if err == nil {
					ext := filepath.Ext(href)
					out.Write([]byte(fmt.Sprintf("data:%s;base64,", strings.Split(mime.TypeByExtension(ext), ";")[0])))
					bw := base64.NewEncoder(base64.StdEncoding, out)
					io.Copy(bw, gres)
					bw.Close()
					res.Close()
				}
			} else {
				out.Write(aVal)
			}
		} else {
			out.Write(aVal)
		}
		out.Write([]byte(`"`))
		if !moreAttr {
			break
		}
	}
}

func attributeHasResource(name, val []byte) bool {
	match := false
	if bytes.Equal(name, []byte("img")) && bytes.Equal(name, []byte("src")) {
		match = true
	}
	if bytes.Equal(name, []byte("link")) && bytes.Equal(name, []byte("href")) {
		match = true
	}
	if bytes.Equal(name, []byte("iframe")) && bytes.Equal(name, []byte("src")) {
		match = true
	}
	if match && bytes.HasPrefix(val, []byte("../../resources/")) {
		return true
	}
	return false
}

func deleteSnapshot(c *gin.Context) {
	u, _ := c.Get("user")
	session := sessions.Default(c)
	defer func() {
		_ = session.Save()
	}()
	bid := c.PostForm("bid")
	sid := c.PostForm("sid")
	if bid == "" || sid == "" {
		return
	}
	var s *model.Snapshot
	err := model.DB.
		Model(&model.Snapshot{}).
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.id = ? and snapshots.bookmark_id = ? and bookmarks.user_id", sid, bid, u.(*model.User).ID).First(&s).Error
	if err != nil {
		setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Snapshot deleted", true)
	}
	if s != nil {
		err = model.DB.Delete(&model.Snapshot{}, "id = ? and bookmark_id = ?", sid, bid).Error
		if err != nil {
			setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
		}
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}
