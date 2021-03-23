window.addEventListener('DOMContentLoaded', async () => {
  const maxEmbeddedFileSize = 102400;

  const textElements = document.querySelectorAll("code[data-path]");
  await Promise.all([...textElements].map(async (textElement) => {
    const size = textElement.getAttribute("data-size") - 0;
    if (maxEmbeddedFileSize < size) {
      textElement.innerHTML = `<span class="file-error">file size is too big to embed. please download from above link to see.</span>`;
      return;
    }
    const url = textElement.getAttribute("data-path");
    const res = await fetch(url);
    const body = await res.text();
    textElement.textContent = body;
    await Prism.highlightElement(textElement, true);
  }));
});
