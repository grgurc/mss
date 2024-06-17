import librosa
import sys
from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import soundfile as sf

if __name__ == "__main__":
    uploads_path = sys.argv[1]
    original_path = Path(uploads_path) / "original.wav"
    
    y, sr = librosa.load(str(original_path))
    D = librosa.stft(y)
    H, P = librosa.decompose.hpss(D)
    harmonic_path = Path(uploads_path) / "median" / "harmonic.wav"
    percussive_path = Path(uploads_path) / "median" / "percussive.wav"
    sf.write(str(harmonic_path), librosa.istft(H), sr)
    sf.write(str(percussive_path), librosa.istft(P), sr)
