from flask import Flask, request, jsonify
import random

app = Flask(__name__)

@app.route('/recommend', methods=['POST'])
def recommend():
    data = request.json
    user_id = data.get("user_id")
    posts = data.get("posts")

    if not user_id or not posts:
        return jsonify({"error": "Invalid input"}), 400

    # Generate recommendations
    scored_posts = [{"id": post["id"], "score": random.random()} for post in posts]
    scored_posts.sort(key=lambda x: x["score"], reverse=True)
    recommended_post_ids = [post["id"] for post in scored_posts[:5]]

    return jsonify({"recommended_posts": recommended_post_ids})

if __name__ == '__main__':
    app.run(port=5000)
