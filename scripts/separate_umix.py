# import openunmix



# separator = openunmix.umxl()
# audio = openunmix.utils.preprocess()
# estimates = separator("a")


import numpy as np
import torchaudio

from scipy.io import wavfile
from openunmix.predict import separate

torchaudio.load()

# Load the WAV file
sample_rate, data = wavfile.read('/Users/grgurcrnogorac/Projects/private/mss/uploads/original.wav')
f = open('/Users/grgurcrnogorac/Projects/private/mss/uploads/original.wav', 'rb')
ten
estimates = separate(f)
