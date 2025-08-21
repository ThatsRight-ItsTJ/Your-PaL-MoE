#!/usr/bin/env python3
"""
Pollinations AI integration script
Free AI image and text generation via Pollinations API
"""

import sys
import json
import requests
from typing import Dict, Any

def main():
    """Main entry point for the script"""
    try:
        # Read request data from stdin
        request_data = json.loads(sys.stdin.read())
        
        # Extract parameters
        prompt = request_data.get('prompt', '')
        model = request_data.get('model', 'openai')
        max_tokens = request_data.get('max_tokens', 100)
        task_type = request_data.get('task_type', 'chat_completion')
        
        if task_type == 'image_generation':
            response = generate_image(prompt, model)
        else:
            response = complete_text(prompt, model, max_tokens)
        
        # Return response in standard format
        result = {
            'success': True,
            'data': response,
            'cost': 0.0,  # Pollinations is free
            'provider': 'pollinations'
        }
        
        print(json.dumps(result))
        
    except Exception as e:
        error_result = {
            'success': False,
            'error': str(e),
            'provider': 'pollinations'
        }
        print(json.dumps(error_result))
        sys.exit(1)

def complete_text(prompt: str, model: str, max_tokens: int) -> Dict[str, Any]:
    """Complete text using Pollinations text API"""
    
    headers = {
        'Content-Type': 'application/json',
        'User-Agent': 'Intelligent-AI-Gateway/1.0'
    }
    
    # Pollinations text endpoint
    url = 'https://text.pollinations.ai/openai'
    
    payload = {
        'messages': [{'role': 'user', 'content': prompt}],
        'model': model,
        'max_tokens': max_tokens,
        'stream': False
    }
    
    try:
        response = requests.post(url, headers=headers, json=payload, timeout=30)
        response.raise_for_status()
        
        data = response.json()
        
        return {
            'text': data.get('choices', [{}])[0].get('message', {}).get('content', ''),
            'model': model,
            'usage': data.get('usage', {
                'prompt_tokens': len(prompt.split()),
                'completion_tokens': max_tokens,
                'total_tokens': len(prompt.split()) + max_tokens
            })
        }
        
    except requests.exceptions.RequestException as e:
        raise Exception(f'Pollinations text API request failed: {e}')

def generate_image(prompt: str, model: str) -> Dict[str, Any]:
    """Generate image using Pollinations image API"""
    
    # Pollinations image endpoint
    base_url = 'https://image.pollinations.ai/prompt'
    
    # Simple GET request for image generation
    image_url = f"{base_url}/{requests.utils.quote(prompt)}"
    
    try:
        # Test if the image URL is accessible
        response = requests.head(image_url, timeout=10)
        response.raise_for_status()
        
        return {
            'url': image_url,
            'prompt': prompt,
            'model': 'pollinations-image',
            'size': '1024x1024'
        }
        
    except requests.exceptions.RequestException as e:
        raise Exception(f'Pollinations image API request failed: {e}')

if __name__ == '__main__':
    main()