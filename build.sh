#!/bin/bash
APP_VERSION="0.0.9"
ANDROID_NDK_HOME="$HOME/Android/android-ndk-r21e" fyne p --os android/arm64 --app-version "${APP_VERSION}" --name "websocket-chat-${APP_VERSION}" --icon wsc.png --id com.martin.chat.gui
