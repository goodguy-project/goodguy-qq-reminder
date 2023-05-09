package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goodguy-project/goodguy-crawl/v2/handler"
	"github.com/goodguy-project/goodguy-crawl/v2/proto"
	"github.com/goodguy-project/goodguy-crawl/v2/util/errorx"
	"github.com/goodguy-project/goodguy-crawl/v2/util/jsonx"
)

var (
	recentContestAllPlatform = []string{
		"codeforces",
		"atcoder",
		"nowcoder",
		"luogu",
		"leetcode",
		"codechef",
		"acwing",
	}
)

func Gao() {
	var wg sync.WaitGroup
	var contests []*proto.GetRecentContestResponse
	var lock sync.Mutex
	for _, platform := range recentContestAllPlatform {
		platform := platform
		wg.Add(1)
		go func() {
			defer wg.Done()
			recentContest, err := handler.GetRecentContest(context.Background(), &proto.GetRecentContestRequest{
				Platform: platform,
			})
			if err != nil {
				_ = errorx.New(err)
				return
			}
			lock.Lock()
			defer lock.Unlock()
			contests = append(contests, recentContest)
		}()
	}
	wg.Wait()
	for _, contest := range contests {
		for _, c := range contest.RecentContest {
			notice(contest.Platform, c)
		}
	}
}

var (
	needSend     = make(map[string]uint64)
	needSendLock sync.Mutex
)

func noticeWhen(timer <-chan time.Time, key string, value uint64, platform string, contest *proto.GetRecentContestResponse_Contest) {
	select {
	case <-timer:
		func() {
			needSendLock.Lock()
			defer needSendLock.Unlock()
			if needSend[key] == value {
				doNotice(platform, contest)
				delete(needSend, key)
			}
		}()
	}
}

func notice(platform string, contest *proto.GetRecentContestResponse_Contest) {
	if contest.Url == "" {
		return
	}
	url := contest.Url
	begin := time.Unix(contest.Timestamp, 0)
	now := time.Now()
	diff := begin.Sub(now) - time.Hour
	if begin.After(now) && diff > 0 && diff < 2*time.Hour {
		needSendLock.Lock()
		defer needSendLock.Unlock()
		timer := time.After(diff)
		random := rand.Uint64()
		needSend[url] = random
		go noticeWhen(timer, url, random, platform, contest)
	}
}

type sendGroupMsgRequest struct {
	GroupId    int64  `json:"group_id,omitempty"`    // 群号
	Message    string `json:"message,omitempty"`     // 要发送的内容
	AutoEscape bool   `json:"auto_escape,omitempty"` // 消息内容是否作为纯文本发送（即不解析CQ码） 只在 message 字段是字符串时有效
}

func sendGroupMsg(req *sendGroupMsgRequest) error {
	request, err := http.NewRequest("POST", "http://qq:5700/send_group_msg", bytes.NewBuffer(jsonx.Marshal(req)))
	if err != nil {
		return errorx.New(err)
	}
	request.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(request)
	if err != nil {
		return errorx.New(err)
	}
	return nil
}

func doNotice(platform string, contest *proto.GetRecentContestResponse_Contest) {
	startTime := time.Unix(contest.GetTimestamp(), 0)
	duration := time.Duration(contest.GetDuration()) * time.Second
	message := fmt.Sprintf("比赛提醒：\n平台：%s\n名称：%s\n时间：%s\n时长：%s\n链接：%s\n",
		platform, contest.GetName(), startTime.String(), duration.String(), contest.GetUrl())
	groupIds := strings.Split(os.Getenv("SEND_GROUP_ID"), ",")
	for _, group := range groupIds {
		if group == "" {
			continue
		}
		group, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			_ = errorx.New(err)
			continue
		}
		err = sendGroupMsg(&sendGroupMsgRequest{
			GroupId:    group,
			Message:    message,
			AutoEscape: false,
		})
		if err != nil {
			_ = errorx.New(err)
		}
	}
}
