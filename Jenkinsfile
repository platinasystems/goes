#!groovy

import groovy.transform.Field

@Field String email_to = 'sw@platinasystems.com'
@Field String email_cc = 'donn@platinasystems.com'
@Field String email_from = 'jenkins-bot@platinasystems.com'
@Field String email_reply_to = 'no-reply@platinasystems.com'

pipeline {
    agent any
    stages {
	stage('Checkout') {
	    steps {
		echo "Running build #${env.BUILD_ID} on ${env.JENKINS_URL}"
		// Change to usual go subdir.
		dir('/home/jenkins/workspace/go/src/github.com/platinasystems/go') {
		    git url: 'https://github.com/platinasystems/go.git'
		}
		dir('/home/jenkins/workspace/go/src/github.com/platinasystems/fe1') {
		    git([
			url: 'git@github.com:platinasystems/fe1.git',
			credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			branch: 'master'
		    ])
		}
		dir('/home/jenkins/workspace/go/src/github.com/platinasystems/firmware-fe1a') {
		    git([
			url: 'git@github.com:platinasystems/firmware-fe1a.git',
			credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			branch: 'master'
		    ])
		}
	    }
	}

	stage('Build') {
	    steps {
		dir('/home/jenkins/workspace/go/src/github.com/platinasystems/go') {
		    echo "Building goes..."
		    sh 'env PATH=/usr/local/go/bin:${PATH} GOPATH=/home/jenkins/workspace/go make noplugin=yes -B goes-platina-mk1-installer'
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
		cc: email_cc,
		to: email_to
	}

	failure {
	    mail body: "GOES build error: ${env.BUILD_URL}",
		from: email_from,
		replyTo: email_reply_to,
		subject: 'GOES BUILD FAILED',
		cc: email_cc,
		to: email_to
	}
    }
}
