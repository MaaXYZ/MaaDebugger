from numpy import ndarray
from PIL import Image


def cvmat_to_image(cvmat: ndarray) -> Image.Image:
    pil = Image.fromarray(cvmat)
    b, g, r = pil.split()
    return Image.merge("RGB", (r, g, b))


def rgb_to_rbg(arr: ndarray) -> ndarray:
    """RGB -> BGR 转换"""
    if arr.ndim == 3 and arr.shape[2] >= 3:
        return arr[:, :, ::-1].copy()
    else:
        return arr
