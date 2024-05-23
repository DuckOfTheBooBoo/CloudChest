import axios from "axios";

export default async function checkTokenValidation(): Promise<boolean> {
    try {
        const response = await axios.post('http://localhost:3000/api/token/check');
        if (response.status == 200) {
            return true;
        }
        return false;
    } catch (error) {
        console.error('Error checking token validation: ', error);
        return false;
    }
}