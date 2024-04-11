import librosa
import matplotlib
import matplotlib.pyplot as plt
import numpy as np
import sys

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Missing argument")
        sys.exit(1)
    file_path = sys.argv[1]
    y, sr = librosa.load(file_path)
    D = librosa.stft(y)
    matplotlib.use('Agg')
    plt.figure(figsize=(10, 4))
    librosa.display.specshow(librosa.amplitude_to_db(np.abs(D), ref=np.max), y_axis='log', x_axis='time')
    plt.colorbar(format='%+2.0f dB')
    plt.tight_layout()
    spectrogram_filename = file_path.rsplit('.', 1)[0] + '_spectrogram.png'
    plt.savefig(spectrogram_filename)
    plt.close()
