#!/bin/bash

# Mute audio
wpctl set-mute @DEFAULT_AUDIO_SINK@ 1

sudo /opt/iceslab/iceslab -u b
