#!/bin/bash

function getAllStepVars {
  ( set -o posix ; set ) | grep GIT | sed -E 's/=.+//g' | xargs
}

function sanitizeOutput {
  echo "$@" | sed -E 's_(.+://).+@_\1oauth-token@_g'
}

function s_info {
  echo "$(sanitizeOutput $@)"
}

function s_success {
  s_info "SUCCESS:$(sanitizeOutput $@)"
}

function s_debug {  
  if [ "${GIT_DEBUG}" == "true" ]; then
    s_info "DEBUG: $(sanitizeOutput $@)"
  fi
}

function s_warning {
  s_info "WARNING: $(sanitizeOutput $@)"
}

function s_fail { 
  echo "ERROR: $(sanitizeOutput $@)"; 
  exit 1;
}

function s_setMessage {
  setMessage "$(sanitizeOutput $@)"
}

# RETURNS REPO_PATH SET in GIT or current WERCKER
function getRepoPath { 
  echo ${GIT_REPO}
}
 

#RETURNS FULL REMOTE PATH OF THE REPO
function getRemoteURL { 
  repo=$(getRepoPath) 
  if [ -n "${GIT_TOKEN}" ]; then
    echo "https://${GIT_USER}:${GIT_TOKEN}@github.com/${repo}" 
  elif [ -n "${repo}" ]; then 
    echo "https://github.com/${repo}"
  else
    echo ""
  fi
}
 
#RETURNS BRANCH WE WANT TO PUSH TO
function getBranch {  
   echo "${GIT_REMOTE_BRANCH}" 
}

#RETURNS BASE DIR WE WANT TO PUSH FROM
function getBaseDir { 
  # if directory provided, cd to it
  if [ -n "$GIT_BASEDIR" ]; then
    echo $GIT_BASEDIR
  else
    echo $(pwd)/
  fi

}
function pullBranch { 
  echo $(git remote rm origin)
  echo $(git remote add origin $1)
   result= echo $(git pull origin $2:$3 2>&1) 
   if [[ $? -ne 0 ]]; then
    s_warning "$result"
    s_fail "failed pulling from $1:$2 to $3"
  else
    s_info "$result"
    s_success "pulled from $1:$2 to $3"
  fi
}
function pushBranch {
  result="$(git push -q -f $1 $2:$3 2>&1)"
  if [[ $? -ne 0 ]]; then
    s_warning "$result"
    s_fail "failed pushing to $3 on $1"
  else
    s_success "pushed to $3 on $1"
  fi
}
 