#!groovy

import groovy.transform.Field

@Field String email_to = 'sw@platinasystems.com'
@Field String email_from = 'jenkins-bot@platinasystems.com'
@Field String email_reply_to = 'no-reply@platinasystems.com'

pipeline {
    agent any
    stages {
	stage('Checkout') {
	    steps {
		echo "Running build #${env.BUILD_ID} on ${env.JENKINS_URL}"
		dir('go') {
		    git([
			url: 'git@github.com:platinasystems/go.git',
			credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			branch: 'master'
			])
		}
		dir('fe1') {
		    git([
			url: 'git@github.com:platinasystems/fe1.git',
			credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			branch: 'master'
			])
		}
		dir('firmware-fe1a') {
		    git([
			url: 'git@github.com:platinasystems/firmware-fe1a.git',
			credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			branch: 'master'
			])
		}
		dir('system-build') {
		    checkout([$class: 'GitSCM',
         	    		      branches: [[name: '*/master']],
         	    		      doGenerateSubmoduleConfigurations: false,
	 	    		      extensions: [[$class: 'SubmoduleOption',
		    		      		  disableSubmodules: false,
						  parentCredentials: true,
						  recursiveSubmodules: true,
						  reference: '',
						  trackingSubmodules: true]],
				      submoduleCfg: [],
				      userRemoteConfigs: [[credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
				      url: 'git@github.com:platinasystems/system-build.git']]])
		}
	    }
	}
	stage('Build') {
	    steps {
		dir('go') {
		    sshagent(credentials: ['570701f7-c819-4db2-bd31-a0da8a452b41']) {
			echo "Updating worktrees"
			sh 'set -x;env;pwd;[ -d worktrees ] && for repo in worktrees/*/*; do echo $repo; [ -d "$repo" ] && (cd $repo;git fetch origin;git reset --hard HEAD;git rebase origin/master);done || true'
			echo "Setting git config"
			sh 'git config --global url.git@github.com:.insteadOf \"https://github.com/\"'
			echo "Building goes..."
			sh 'env PATH=/usr/local/go/bin:/usr/local/x-tools/arm-unknown-linux-gnueabi/bin:${PATH} go run ./main/goes-build/main.go -x -v -z'
		    }
		}
	    }
	}
    }

    post {
	success {
	    mail body: "GOES build ok: ${env.BUILD_URL}\n\ngoes-platina-mk1-installer is stored on platina4 at /home/jenkins/workspace/go/src/github.com/platinasystems/go/goes-platina-mk1\neg.\nscp 172.16.2.23:/home/jenkins/workspace/go/src/github.com/platinasystems/go/goes-platina-mk1 ~/path/to/somewhere/",
		from: email_from,
		replyTo: email_reply_to,
		subject: 'GOES build ok',
		to: email_to
	}

	failure {
		cleanWs()
		mail body: "GOES build error: ${env.BUILD_URL}",
		from: email_from,
		replyTo: email_reply_to,
		subject: 'GOES BUILD FAILED',
		to: email_to
	}
    }
}
