function toggleReferences(references) {
    if (references.classList.contains('hide')) {
        references.classList.remove('hide');
    } else {
        references.classList.add('hide');
    }
}

function isQuestionCardDisplayingReferences(questionCard) {
    return !questionCard.getElementsByClassName('references')[0].classList.contains('hide');
}

document.querySelectorAll('.single-reference-toggle').forEach(toggle => {
    const questionCard = toggle.closest('.question-card');
    const references = questionCard.getElementsByClassName('references')[0];
    toggle.addEventListener('click', function() {
        toggleReferences(references);
    });
});

document.getElementById('toggle-reference-button').addEventListener('click', function() {
    const references = document.querySelectorAll('.references');
    const isGlobalReferencesToggled = this.classList.contains('toggled');
    if (isGlobalReferencesToggled) {
        this.classList.remove('toggled');
    } else {
        this.classList.add('toggled');
    }
    references.forEach(reference => {
        const questionCard = reference.closest('.question-card');
        if (isGlobalReferencesToggled === isQuestionCardDisplayingReferences(questionCard)) {
            toggleReferences(reference);
        }
    });
});

const CORRECT_ANSWER = 'correct';
const WRONG_ANSWER = 'wrong';
const MISSING_ANSWER = 'missing';

function toggleAnswer(questionCard, displayAnswer) {
    const correctAnswers = questionCard.dataset.correct.split(',');
    const choices = questionCard.querySelectorAll('.choice input[type="checkbox"]');

    choices.forEach(choice => {
        const parentLabel = choice.parentElement;
        if (displayAnswer) {
            if (choice.checked && correctAnswers.includes(choice.value)) {
                parentLabel.classList.add(CORRECT_ANSWER);
                parentLabel.classList.remove(WRONG_ANSWER, MISSING_ANSWER);
            } else if (choice.checked && !correctAnswers.includes(choice.value)) {
                parentLabel.classList.add(WRONG_ANSWER);
                parentLabel.classList.remove(CORRECT_ANSWER, MISSING_ANSWER);
            } else if (!choice.checked && correctAnswers.includes(choice.value)) {
                parentLabel.classList.add(MISSING_ANSWER);
                parentLabel.classList.remove(CORRECT_ANSWER, WRONG_ANSWER);
            } else {
                parentLabel.classList.remove(CORRECT_ANSWER, WRONG_ANSWER, MISSING_ANSWER);
            }
        } else {
            parentLabel.classList.remove(CORRECT_ANSWER, WRONG_ANSWER, MISSING_ANSWER);
        }
    });
}

function isQuestionCardDisplayingAnswer(questionCard) {
    const choices = questionCard.querySelectorAll('.choice input[type="checkbox"]');
    return Array.from(choices).some(choice =>
        [CORRECT_ANSWER, WRONG_ANSWER, MISSING_ANSWER].some(answer =>
            choice.parentElement.classList.contains(answer)
        )
    );
}

document.querySelectorAll('.single-answer-toggle').forEach(toggle => {
    const questionCard = toggle.closest('.question-card');
    toggle.addEventListener('click', function() {
        const isDisplayingAnswers = isQuestionCardDisplayingAnswer(questionCard);
        toggleAnswer(questionCard, !isDisplayingAnswers);
        if (isDisplayingAnswers) {
            this.innerHTML = '<i class="fas fa-eye"></i>';
        } else {
            this.innerHTML = '<i class="fas fa-eye-slash"></i>';
        }
    });
});

document.getElementById('toggle-button').addEventListener('click', function() {
    const isGlobalAnswerToggled = this.className.includes('toggled');
    const allSingleAnswerToggles = document.querySelectorAll('.single-answer-toggle');
    allSingleAnswerToggles.forEach(toggle => {
        const questionCardAnswerToggled = isQuestionCardDisplayingAnswer(toggle.closest('.question-card'));
        if (questionCardAnswerToggled === isGlobalAnswerToggled) {
            toggle.click();
        }
    });

    if (isGlobalAnswerToggled) {
        this.innerHTML = '<i class="fas fa-eye"></i>';
        this.classList.remove('toggled');
    } else {
        this.innerHTML = '<i class="fas fa-eye-slash"></i>';
        this.classList.add('toggled');
    }
    saveState();
});

const CHOICE_CHECK_MAP_KEY = 'choiceCheckMap';
const READ_CHECK_MAP_KEY = 'readCheckMap';
const SHOW_ANSWERS_KEY = 'showAnswers';

// Load state from localStorage
function loadState() {
    const choiceCheckboxes = document.querySelectorAll('.choice input[type="checkbox"]');
    const choiceCheckMap = JSON.parse(localStorage.getItem(CHOICE_CHECK_MAP_KEY)) || {};
    choiceCheckboxes.forEach(checkbox => {
        const isChecked = choiceCheckMap[`${checkbox.name}-${checkbox.value}`];
        if (isChecked) {
            checkbox.checked = true;
        }
    });

    const readCheckboxes = document.querySelectorAll('.read-checkbox');
    const readCheckMap = JSON.parse(localStorage.getItem(READ_CHECK_MAP_KEY)) || {};
    readCheckboxes.forEach(checkbox => {
        const isChecked = readCheckMap[checkbox.closest('.question-card').id];
        if (isChecked) {
            checkbox.checked = true;
            checkbox.closest('.question-card').classList.add('read');
        }
    });

    const isShowingAnswers = localStorage.getItem(SHOW_ANSWERS_KEY) === 'true';
    if (isShowingAnswers) {
        document.getElementById('toggle-button').click();
    }
}

// Save state to localStorage
function saveState() {
    const choiceCheckboxes = document.querySelectorAll('.choice input[type="checkbox"]');
    const choiceCheckMap = {};
    choiceCheckboxes.forEach(checkbox => {
        choiceCheckMap[`${checkbox.name}-${checkbox.value}`] = checkbox.checked;
    });
    localStorage.setItem(CHOICE_CHECK_MAP_KEY, JSON.stringify(choiceCheckMap));

    const readCheckboxes = document.querySelectorAll('.read-checkbox');
    const readCheckMap = {};
    readCheckboxes.forEach(checkbox => {
        readCheckMap[checkbox.closest('.question-card').id] = checkbox.checked;
    });
    localStorage.setItem(READ_CHECK_MAP_KEY, JSON.stringify(readCheckMap));

    const isShowingAnswers = document.getElementById('toggle-button').className.includes('toggled');
    localStorage.setItem(SHOW_ANSWERS_KEY, `${isShowingAnswers}`);
}

document.addEventListener('DOMContentLoaded', loadState);

document.querySelectorAll('.choice input[type="checkbox"]').forEach(checkbox => {
    checkbox.addEventListener('change', saveState);
});

document.querySelectorAll('.read-checkbox').forEach(checkbox => {
    checkbox.addEventListener('change', function() {
        const card = this.closest('.question-card');
        if (this.checked) {
            card.classList.add('read');
        } else {
            card.classList.remove('read');
        }
        saveState();
    });
});

document.getElementById('menu-button').addEventListener('click', function() {
    const menuItems = document.getElementById('menu-items');
    menuItems.style.display = menuItems.style.display === 'block' ? 'none' : 'block';
    if (menuItems.style.display === 'block') {
        menuItems.focus();
    }
});

document.addEventListener('click', function(event) {
    const menuItems = document.getElementById('menu-items');
    const menuButton = document.getElementById('menu-button');
    if (menuItems.style.display === 'block' && !menuButton.contains(event.target) && !menuItems.contains(event.target)) {
        menuItems.style.display = 'none';
    }
});

document.getElementById('clear-button').addEventListener('click', function() {
    const choiceCheckboxes = document.querySelectorAll('.choice input[type="checkbox"]');
    choiceCheckboxes.forEach(checkbox => {
        checkbox.checked = false;
    });
    localStorage.setItem(CHOICE_CHECK_MAP_KEY, JSON.stringify({}));

    // Optionally, clear the answer highlights if they are visible
    const isShowingAnswers = localStorage.getItem(SHOW_ANSWERS_KEY) === 'true';
    const allQuestionCards = document.querySelectorAll('.question-card');
    allQuestionCards.forEach(card => {
        const choices = card.querySelectorAll('.choice input[type="checkbox"]');
        const correctAnswers = card.dataset.correct.split(',');
        choices.forEach(choice => {
            const parentLabel = choice.parentElement;
            parentLabel.classList.remove('missing', 'wrong', 'correct');
            if (isShowingAnswers && correctAnswers.includes(choice.value)) {
                parentLabel.classList.add('missing');
            }
        });
    });
});

function scrollToQuestion(questionId) {
    document.getElementById(questionId).scrollIntoView();
    document.getElementById('menu-items').style.display = 'none';
}
