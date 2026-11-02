document$.subscribe(() => {
    document.querySelectorAll('.md-content code').forEach((code) => {
        if (!code.closest('pre, a'))
            code.setAttribute('data-clipboard-text', code.textContent)
    })
})
