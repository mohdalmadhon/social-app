import { LIMITS, charCount } from '@/helpers/limits';
import { isLinkField, validateLink } from '@/helpers/links';

export function validateAboutField(value = '', field = '') {
    if (isLinkField(field)) {
        return validateLink(field, value);
    }

    if (charCount(value) > LIMITS.aboutField) {
        return `Must not exceed ${LIMITS.aboutField} characters`;
    }

    return '';
}
