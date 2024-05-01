import librosa
import matplotlib
import sys

import matplotlib.pyplot as plt
import numpy as np
import soundfile as sf

if __name__ == "__main__":
    file_path = "./uploads/original.wav"
    y, sr = librosa.load(file_path)
    D = librosa.stft(y)
    H, P = librosa.decompose.hpss(D)
    sf.write("./uploads/median_harmonic.wav", librosa.istft(H), sr)
    sf.write("./uploads/median_percussive.wav", librosa.istft(P), sr)
