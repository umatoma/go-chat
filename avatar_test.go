package main

import (
	"testing"
	"path/filepath"
	"io/ioutil"
	"os"
)

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	testChatUser := &chatUser{
		userInfo: map[string]string{},
	}
	url, err := authAvatar.GetAvatarURL(testChatUser)
	if err != ErrNoAvatarURL {
		t.Error("値が存在しない場合、AuthAvatar.GetAvatarURLはErrNoAvatarURLを返すべきです")
	}

	testUrl := "http://url-to-avatar/"
	testChatUser = &chatUser{
		userInfo: map[string]string{"avatar_url": testUrl},
	}
	url, err = authAvatar.GetAvatarURL(testChatUser)
	if err != nil {
		t.Error("値が存在する場合、AuthAvatar.GetAvatarURLはエラーを返すべきではありません")
	} else {
		if url != testUrl {
			t.Error("AuthAvatar.GetAvatarURLは正しいURLを返すべきです")
		}
	}
}

func TestGravatarAvatar (t *testing.T) {
	var gravatarAvatar GravatarAvatar
	user := &chatUser{uniqueID: "abc"}
	url, err := gravatarAvatar.GetAvatarURL(user)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURLはエラーを返すべきではありません")
	}
	if url != "//www.gravatar.com/avatar/abc" {
		t.Errorf("GravatarAvatar.GetAvatarURLが%sという誤った値を返しました", url)
	}
}

func TestFileSystemAvatar (t *testing.T) {
	filename := filepath.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename, []byte{}, 0777)
	defer func() { os.Remove(filename) }()

	var fileSystemAvatar FileSystemAvatar
	user := &chatUser{uniqueID: "abc"}
	url, err := fileSystemAvatar.GetAvatarURL(user)
	if err != nil {
		t.Error("FileSystemAvatar.GetAvatarURLはエラーを返すべきではありません")
	}
	if url != "/avatars/abc.jpg" {
		t.Errorf("FileSystemAvatar.GetAvatarURLが%sという誤った値を返しました", url)
	}
}
