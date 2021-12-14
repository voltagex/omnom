function addTag(event, tagChipContainer, tags) {
    const value = event.target.value;
    renderTag(value, tagChipContainer, tags);
    tags.push(value);
}

function renderTag(value, tagChipContainer, tags) {
    const tagTemplate = `<div class="control chip-control">
        <span class="tag is-rounded">
            ${value}
            <button type="button" class="delete is-small"></button>
        </span>
    </div>`
    const template = document.createElement('template');
    template.innerHTML = tagTemplate;
    const deleteButton = template.content.querySelector('button');
    deleteButton.addEventListener('click', deleteTag.bind({}, template.content.firstChild, tagChipContainer, tags));
    tagChipContainer.appendChild(template.content.firstChild);
}

function deleteTag(chipElement, tagChipContainer, tags) {
    tagChipContainer.removeChild(chipElement);
    tags = [...tagChipContainer.children].map(child => child.innerText);
}

function renderTags(tags, tagChipContainer) {
    const fragment = document.createDocumentFragment();
    tags.forEach(tag => {
        renderTag(tag, fragment);
    });
    tagChipContainer.appendChild(fragment);
}

function TagInputController(inputElement, chipContainer) {
    let tags = [];
    this.getTags = () => (tags);
    this.renderTags = () => renderTags(tags, chipContainer);
    inputElement?.addEventListener('change', (event) => { addTag(event, chipContainer, tags); inputElement.value = '' });
}

export { TagInputController };