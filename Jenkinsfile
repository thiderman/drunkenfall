pipeline {
  agent any

  stages {
    // Setup steps
    stage('Go dependencies') {
      steps {
        sh 'make download'
      }
    }
    stage('npm dependencies') {
      steps {
        sh 'make npm'
      }
    }

    // Verification steps
    stage('Unit tests') {
      steps {
        sh 'make test'
      }
    }
    stage('Linting checks') {
      steps {
        sh 'make lint'
      }
    }
    stage('Race condition check') {
      steps {
        sh 'make race'
      }
    }

    // Building step
    stage('Compile application') {
      steps {
        sh 'make drunkenfall'
      }
    }
    stage('Build npm') {
      steps {
        sh 'make npm-dist'
      }
    }

    // Deploy step
    stage('Deploy') {
      steps {
        echo 'make deploy'
      }
    }
  }
}