#!/bin/sh   

# 加载方法
DIR=$( cd "$( dirname $0 )" && pwd )
. ${DIR}/functions.sh

# 不允许无Token
if [ ! -n "$GIT_TOKEN" ]; then
  s_warning "Your gh_token may be compromised."
fi 

# 获取Git repo 名
repo=$(getRepoPath)
if [ ! -n "$repo" ]; then 
  s_fail "env GIT_REPO not found." 
fi
s_info "using github repo \"$repo\""

remoteURL=$(getRemoteURL) 
if [ ! -n "$remoteURL" ]; then 
  s_fail "env GIT_REPO not found." 
fi
s_info "remote URL will be $remoteURL"

targetDir=$(getBaseDir)

# setup branch
remoteBranch=$(getBranch)

localBranch=$GIT_LOCAL_BRANCH
s_info "git local will be $targetDir" 

# 判断本地项目目录是否存在
if [ ! -d "$targetDir" ]; then 
  s_fail "git local directory doesn't exist."
fi

cd $targetDir 
s_info "current pwd is $(pwd)" 

git config user.email $GIT_USER_EMAIL
git config user.name $GIT_USER

git remote rm origin
git remote add origin $remoteURL
git checkout $localBranch
git pull origin $remoteBranch 
#git add --all . > /dev/null
if [ ! -n "$COMMIT_FILE" ]; then
  git add --all . > /dev/null 
else 
    git add $COMMIT_FILE > /dev/null
fi 
git status 

if git diff --cached --exit-code --quiet; then
  s_success "Nothing changed. We do not need to push"
else
  git commit -am "$COMMIT_MSG"  --allow-empty > /dev/null
  s_info "will pull to : $remoteURL localBranch=$localBranch remoteBranch=$remoteBranch"
  pushBranch $remoteURL $localBranch $remoteBranch 
fi 