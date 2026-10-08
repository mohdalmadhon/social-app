
export function checkSessionResponse(resp) {
    if(resp.status === 401 ) {
        return false;
    }
    return true
}