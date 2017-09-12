package rebot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/ysqi/com"

	"golang.org/x/oauth2"

	"github.com/spf13/viper"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	repoOwner            = "ysqi"
	repoName             = "goall"
	commentPrefix        = "From Rebot\n" //发送Comment时将包含此前缀，用于标记
	approveCommentPrefix = "approve:"     //comment 确认审批
	pageLabelZone        = []string{"article", "book", "job", "news", "pkg"}
)

// CommentMsg Comment信息
type CommentMsg struct {
	Owner    string
	Repo     string
	Body     string
	Issue    int
	TryCount int
}

// GithubWorker Github相关处理
type GithubWorker struct {
	webhookSecretKey []byte
	eventChan        chan interface{}
	exitChan         chan bool

	client *github.Client
	owner  string
	repo   string
}

// NewGithubWorker 创建实例，必须存在密钥
func NewGithubWorker() (*GithubWorker, error) {
	key := viper.GetString("GitHub_Webhook_Key")
	if key == "" {
		return nil, errors.New("环境变量缺失GitHub_Webhook_Key")
	}

	max := viper.GetInt("Github_Webhook_MaxWorker")
	if max == 0 {
		max = 1000
	}

	s := &GithubWorker{
		webhookSecretKey: []byte(key),
		eventChan:        make(chan interface{}, max),
		exitChan:         make(chan bool),
		owner:            repoOwner,
		repo:             repoName,
	}
	//create github client
	tc := oauth2.NewClient(context.Background(),
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: viper.GetString("GitHub_Token"),
			},
		))
	s.client = github.NewClient(tc)

	return s, nil
}

// GithubHookHandler Github Web Hook回调处理入口
func (s *GithubWorker) GithubHookHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, s.webhookSecretKey)

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		glog.Error("Github Hook:", err)
		outputError(w, "ERR")
		return
	}
	glog.Infoln("New github event:", reflect.TypeOf(event))

	s.eventChan <- event

	w.Write([]byte("OK"))
}

// StartWorker 启动Event监控处理
func (s *GithubWorker) StartWorker() { go s.startWorker() }
func (s *GithubWorker) startWorker() {
	closed := false
	for {
		select {
		case <-s.exitChan:
			closed = true
		case event, ok := <-s.eventChan:
			if !ok {
				closed = true
				break
			}
			if event == nil {
				break
			}
			eventTyp := reflect.TypeOf(event).String()
			glog.Infoln("begin process message", eventTyp)

			switch v := event.(type) {
			case CommentMsg:
				s.postComment(v)
			case *github.PingEvent:
				glog.Infoln("Ping Event:", github.Stringify(v))
			case *github.IssuesEvent:
				s.processIssueEvent(v)
			case *github.IssueEvent:
				// s.processIssueEvent(v)
			case *github.IssueCommentEvent:
				s.processIssCommentEvent(v)
			default:
				glog.Warningln("unprocess message", eventTyp)
			}
			glog.Infoln("end process message", eventTyp)
		}
		if closed {
			break
		}
	}
}

// checkRepo 检查是否是正确的Repo事件
func (s *GithubWorker) checkRepo(repo *github.Repository) error {
	repoName := repo.GetName()
	if repoName != s.repo {
		return fmt.Errorf("repo want %s,but got %s", s.repo, repoName)
	}
	owner := repo.Owner.GetLogin()
	if owner != s.owner {
		return fmt.Errorf("owner want %s,but got %s", s.owner, owner)
	}
	return nil
}

// processIssueEvent
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#issuesevent
// Triggered when an issue is assigned, unassigned, labeled, unlabeled, opened,
// edited, milestoned, demilestoned, closed, or reopened.
func (s *GithubWorker) processIssueEvent(event *github.IssuesEvent) {
	if err := s.checkRepo(event.Repo); err != nil {
		glog.Warningln("issue event:", event.Action, err)
		return
	}
	//仅处理Opened和Edited消息
	action := event.GetAction()
	if action != "opened" && action != "edited" {
		return
	}
	//如果是Edited则需要检查该Issue是否已处理完毕
	if action == "edited" {

	}

	body := event.Issue.GetBody()
	// 如果是自动内容前缀则忽略
	if strings.HasPrefix(commentPrefix, body) {
		return
	}
	//解析内容，解析正确后发送Comment，等待管理员审批
	page, err := NewPageInfo(body)
	if err != nil {
		s.pushComment(event.Issue, err)
		return
	}
	result := page.ToIni()
	s.pushComment(event.Issue, fmt.Sprintf("提交正确，等待管理员@%s审批，解析后内容如下\n```ini\n%s```", s.owner, result))

}

// processIssCommentEvent
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#issuecommentevent
// Triggered when an issue comment is created, edited, or deleted.
// 这里只处理审批通过消息，审批内容必须是有指定人员审批（目前只有管理员）。
// 当审批人提交审批通过内容时，将该内容创建对应的Markdown内容文件，并提交。
func (s *GithubWorker) processIssCommentEvent(event *github.IssueCommentEvent) {
	if err := s.checkRepo(event.Repo); err != nil {
		glog.Warningln("issue event:", event.Action, err)
		return
	}
	//仅处理created消息
	action := event.GetAction()
	if action != "created" {
		return
	}
	//只处理管理员comment
	if s.owner != event.Comment.User.GetLogin() {
		return
	}
	body := event.Comment.GetBody()
	//检查是否是审批消息
	if !strings.HasPrefix(body, approveCommentPrefix) {
		return
	}
	label := strings.ToLower(strings.TrimPrefix(body, approveCommentPrefix))
	// 检查分类
	if label == "" || !com.IsSliceContainsStr(pageLabelZone, label) {
		s.pushComment(event.Issue, fmt.Sprintf("所配置的label=%q非法，不属有效值:%s，请重新提交。", label, strings.Join(pageLabelZone, ",")))
		return
	}

	//将Issue设置label
	// TODO: 是否有必要将此请求也放入队列？
	_, _, err := s.client.Issues.AddLabelsToIssue(context.Background(), s.owner, s.repo, event.Issue.GetNumber(), []string{label})
	if err != nil {
		s.pushComment(event.Issue, err)
		return
	}

	//解析内容，解析正确后发送Comment，等待管理员审批
	page, err := NewPageInfo(event.Issue.GetBody())
	if err != nil {
		s.pushComment(event.Issue, err)
		return
	}
	content, err := hugo.GenPageContent(label, page, event.Issue)
	if err != nil {
		s.pushComment(event.Issue, err)
		return
	}
	//创建文件
	file, err := hugo.CreatePage(label,fmt.Sprintf("Issue-%d.md", event.Issue.GetNumber()),content)
	if err != nil {
		s.pushComment(event.Issue, "创建md文件失败，"+err.Error())
		return
	}
	//发布
	err=hugo.Commit(file, fmt.Sprintf("new content added create by @%s from #%d", 
		event.Issue.User.GetLogin(), 
		event.Issue.GetNumber()))
	if err != nil {
		s.pushComment(event.Issue, "创建md文件并提交到Git时失败，"+err.Error())
		return
	}
}

func (s *GithubWorker) pushComment(issue *github.Issue, comment interface{}) {
	var body string
	switch v := comment.(type) {
	case error:
		body = fmt.Sprintf("@%s,出现错误:%s", issue.User.GetLogin(), v.Error())
	default:
		body = fmt.Sprintf("%v", v)
	}
	body = commentPrefix + body
	c := CommentMsg{
		Repo:  s.repo,
		Owner: s.owner,
		Issue: issue.GetNumber(),
		Body:  body,
	}
	s.eventChan <- c

}

var postCommentMaxTryCount = 4

// postComment 上传Comment信息到Issue
// 设计为可重复上传，失败后将重新放入eventChan队列中，最多尝试4次。
// TOOD: 非阶梯式重试，有改进空间
func (s *GithubWorker) postComment(comment CommentMsg) {
	input := &github.IssueComment{
		Body: &comment.Body,
	}
	comment.TryCount++
	_, resp, err := s.client.Issues.CreateComment(
		context.Background(),
		comment.Owner, comment.Repo,
		comment.Issue,
		input)
	if err == nil {
		return
	}
	if err != nil {
		// 1xx消息
		// 2xx成功
		// 3xx重定向
		// 4xx客户端错误
		// 5xx服务器错误
		if resp != nil && resp.StatusCode < 500 {
			glog.Errorf("发送失败:%s，放弃上传:%+v\n", err, comment)
			return
		}
		glog.Errorf("发送失败:%s，将重发消息:%+v\n", err, comment)

	}

	if comment.TryCount >= postCommentMaxTryCount {
		glog.Errorf("发送失败:%s。已重试%d，放弃上传:%+v\n", err, comment.TryCount, comment)
		return
	}
	s.eventChan <- comment
}
