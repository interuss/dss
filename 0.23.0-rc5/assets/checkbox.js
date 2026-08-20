// https://github.com/squidfunk/mkdocs-material/discussions/5235

const allLists = Array.from(document.querySelectorAll(".task-list"))
const LOCALSTORAGE_KEY = "checkedItems"

window.addEventListener('load', () => {
    const path = window.location.pathname
    let checkedItems = localStorage.getItem(LOCALSTORAGE_KEY)

    if (!checkedItems)
        checkedItems = []
    else
        checkedItems = JSON.parse(checkedItems)

    allLists.forEach((list, listIndex) => {
        const items = Array.from(list.querySelectorAll(".task-list-item"))
        items.forEach((item) => {
            const text = item.textContent
            const itemKey = `${path}:${listIndex}:${text}`

            if (checkedItems.includes(itemKey)) {
                const checkbox = item.querySelector('input[type="checkbox"]')
                checkbox.checked = true
            }
        })
    })
})

document.body.addEventListener('click', (e) => {
    if (!e.target.matches('.task-list-item label input'))
        return

    const checkbox = e.target
    const checked = checkbox.checked
    const li = checkbox.parentElement.parentElement
    const text = li.textContent
    const list = li.parentElement
    const listIndex = allLists.indexOf(list)
    const path = window.location.pathname

    const itemKey = `${path}:${listIndex}:${text}`

    let checkedItems = JSON.parse(localStorage.getItem(LOCALSTORAGE_KEY))

    if (!checkedItems)
        checkedItems = []

    // Ensure we don't have duplicates by always removing the item first
    checkedItems = checkedItems.filter(i => i !== itemKey)

    if (checked)
        checkedItems.push(itemKey)

    localStorage.setItem(LOCALSTORAGE_KEY, JSON.stringify(checkedItems))
})
