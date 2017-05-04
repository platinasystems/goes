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
		    // Clone private 'fe1' repo to a holding subdir.
		    dir('CheckoutFe1') {
			git([
			    url: 'git@github.com:platinasystems/fe1.git',
			    credentialsId: "570701f7-c819-4db2-bd31-a0da8a452b41",
			    branch: 'master'
			])

			sh('rm -rf ./vnet/devices/ethernet/switch/fe1/')
			sh('mkdir -p ./vnet/devices/ethernet/switch/fe1/')
			sh('cp -r ./CheckoutFe1/* ./vnet/devices/ethernet/switch/fe1')
		    }
		}
	    }
	}

	stage('Build') {
	    environment {
		env.PATH = "/usr/local/go/bin/:${env.PATH}"
		env.GOPATH = "/home/jenkins/workspace/go"
	    }
	    steps {
		dir('/home/jenkins/workspace/go/src/github.com/platinasystems/go') {
		    echo "Building goes..."
		    sh 'make -B goes-platina-mk1'
		}
	    }
	}
    }

    post {
	success {
	    mail body: "GOES build ok: ${env.BUILD_URL}",
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
