#!/bin/sh  

# 加载方法
DIR=$( cd "$( dirname $0 )" && pwd )
. ${DIR}/functions.sh


if [ ! -n "$GIT_TOKEN" ]; then
  s_fail "Your gh_token may be compromised."
fi 

repo=$(getRepoPath)
if [ ! -n "$repo" ]; then 
  s_fail "env GIT_REPO not found." 
fi

s_info "using github repo \"$repo\""

remoteURL=$(getRemoteURL) 
if [ ! -n "$remoteURL" ]; then 
  s_fail "env GIT_USER not found" 
fi
s_info "remote URL will be $remoteURL"

targetDir=$(getBaseDir)

# setup branch
remoteBranch=$(getBranch)
 
localBranch=$GIT_LOCAL_BRANCH

# git clone if dir not exit
s_info "$targetDir"
if [[ ! -d "$targetDir" ]]; then
  s_info "beging exec: git clone -b $remoteBranch $remoteURL ****"
  git clone -b $remoteBranch $remoteURL $targetDir
fi 

cd $targetDir

git checkout $localBranch
git remote rm origin
git remote add origin $remoteURL
# git pull $remoteURL $localBranch
pullBranch  $remoteURL $remoteBranch  $localBranch 

