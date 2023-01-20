#!/usr/bin/env groovy

pipeline {
    agent {
        label 'Slave'
    }

    options {
        timestamps()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    parameters {
        string(name: 'version', defaultValue: 'latest', description: 'App build version')
        string(name: 'github_username', defaultValue: 'praveenprem', description: 'Github registry auth user')
    }

    stages {
        stage('Build image') {
            steps {
                withCredentials([file(credentialsId: '7b372ab2-a105-42f0-996c-eb1d18fb8a8a', variable: 'FILE')]) {
                    sh 'cat $FILE > netrc'
                }
                sh "make docker VERSION=${params.version}"
            }
        }
        stage('Build publish') {
            environment {
                REGISTRY_TOKEN = credentials("eb9b3bdf-cdac-4dd6-a469-ba09eb82ebaa")
            }
            steps {
                sh "make auth publish VERSION=${params.version} REGISTRY_AUTH_USER=${params.github_username}"
            }
        }
    }

    post {
        success {
            slackSend color: "good", message: "${env.JOB_BASE_NAME} - #${env.BUILD_NUMBER} Success - (<${env.BUILD_URL}|Open>)"
        }

        failure {
            slackSend color: "danger", message: "${env.JOB_BASE_NAME} - #${env.BUILD_NUMBER} Failure - (<${env.BUILD_URL}|Open>)"
        }

        cleanup {
            deleteDir()
        }
    }
}
