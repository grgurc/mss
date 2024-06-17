import nussl
import sys
import pathlib

if __name__ == "__main__":
    uploads_path = sys.argv[1]
    input_file = pathlib.Path(uploads_path) / "original.wav"
    audio_signal = nussl.AudioSignal(input_file)

    bg_file = pathlib.Path(uploads_path) / "repet" / "background.wav"
    fg_file = pathlib.Path(uploads_path) / "repet" / "foreground.wav"

    repet = nussl.separation.primitive.RepetSim(audio_signal) # we use repet-sim since it might be better
    bg, fg = repet()

    bg.write_audio_to_file(bg_file)
    fg.write_audio_to_file(fg_file)