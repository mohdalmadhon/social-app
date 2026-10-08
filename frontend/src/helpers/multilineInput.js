import { limitLines } from '@/helpers/limits';

export function isCoarsePointer() {
    return typeof window !== 'undefined'
        && typeof window.matchMedia === 'function'
        && window.matchMedia('(pointer: coarse)').matches;
}

export function resizeTextarea(element, maxHeight = 140) {
    if (!element) {
        return;
    }

    element.style.height = 'auto';

    const height = Math.min(element.scrollHeight, maxHeight);

    element.style.height = `${height}px`;
    element.style.overflowY = element.scrollHeight > maxHeight ? 'auto' : 'hidden';
}

export function resetTextarea(element) {
    if (!element) {
        return;
    }

    element.style.height = '';
    element.style.overflowY = 'hidden';
}

export function handleEnterKey(event, submit) {
    if (event.key !== 'Enter' || event.isComposing || event.keyCode === 229) {
        return;
    }

    if (event.shiftKey || event.ctrlKey || event.altKey || event.metaKey || isCoarsePointer()) {
        return;
    }

    event.preventDefault();
    submit();
}

export function enforceLines(event, maxLines) {
    const element = event.target;
    const limited = limitLines(element.value, maxLines);

    if (limited !== element.value) {
        element.value = limited;
    }

    return element.value;
}
