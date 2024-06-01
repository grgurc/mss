import openunmix
import openunmix.data
import openunmix.predict
import torch
import torchaudio
import sys

from pathlib import Path

if __name__ == "__main__":
    uploads_folder = sys.argv[1]

    input_file = Path(uploads_folder) / "original.wav"

    model = "umxl"
    use_cuda = torch.cuda.is_available()
    device = torch.device("cuda" if use_cuda else "cpu")

    separator = openunmix.utils.load_separator(
        model_str_or_path=model,
        niter=1,
        residual=None,
        wiener_win_len=300,
        device=device,
        pretrained=True,
        filterbank="torch",
    )

    separator.freeze()
    separator.to(device)

    audio, rate = openunmix.data.load_audio(str(input_file))
    estimates = openunmix.predict.separate(
        audio=audio,
        rate=rate,
        aggregate_dict=None,
        separator=separator,
        device=device,
    )

    outdir = Path(uploads_folder) / "unmix" 
    outdir.mkdir(exist_ok=True, parents=True)

    for target, estimate in estimates.items():
        target_path = str(outdir / Path(target).with_suffix(".wav"))
        torchaudio.save(
            target_path,
            torch.squeeze(estimate).to("cpu"),
            sample_rate=separator.sample_rate,
        )