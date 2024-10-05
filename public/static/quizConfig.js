const multiSelectWithoutCtrl = (elemSelector) => {
    let options = [].slice.call(document.querySelectorAll(`${elemSelector} option`));
    options.forEach(function(element) {
        element.addEventListener("mousedown", function(e) {
            e.preventDefault();
            element.parentElement.focus();
            this.selected = !this.selected;
            return false;
        }, false);
    });
}

document.addEventListener('DOMContentLoaded', function() {
    multiSelectWithoutCtrl('#rules-filter');
});

const numberInputs = Array.from(document.getElementsByClassName('number-input'));
numberInputs.forEach(el => el.addEventListener('input', function(e) {
    e.target.value = e.target.value.replace(/[^0-9.]/g, '');
    const numQuestions = parseInt(e.target.value);
    const maxNumQuestions = parseInt(e.target.max);
    if (numQuestions === 0) {
        e.target.value = null;
    }
    if (!isNaN(maxNumQuestions) && numQuestions > maxNumQuestions) {
        e.target.value = maxNumQuestions;
    }
})
)

document.getElementById('rules-filter-select-all').addEventListener('click', function() {
    const select = document.getElementById('rules-filter');
    for (let i = 0; i < select.options.length; i++) {
        select.options[i].selected = true;
    }
});

document.getElementById('rules-filter-clear-all').addEventListener('click', function() {
    const select = document.getElementById('rules-filter');
    for (let i = 0; i < select.options.length; i++) {
        select.options[i].selected = false;
    }
});
