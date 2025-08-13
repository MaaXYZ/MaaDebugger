from numpy import ndarray
from PIL import Image


def cvmat_to_image(cvmat: ndarray) -> Image.Image:
    pil = Image.fromarray(cvmat)
    b, g, r = pil.split()
    return Image.merge("RGB", (r, g, b))
