from flask import Flask, request, jsonify
import pickle
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.preprocessing import MultiLabelBinarizer

app = Flask(__name__)

# Step 1: Load the vectorizer, classifier, and label binarizer
try:
    with open('vectorizer.pkl', 'rb') as f:
        vectorizer = pickle.load(f)
    
    with open('classifier.pkl', 'rb') as f:
        classifier = pickle.load(f)
    
    with open('label_binarizer.pkl', 'rb') as f:
        mlb = pickle.load(f)
    
    print("Model, vectorizer, and label binarizer loaded successfully!")
except Exception as e:
    print(f"Error loading files: {e}")
    raise e

@app.route('/predict', methods=['POST'])
def predict_tags():
    try:
        # Step 2: Get the input content from the request
        content = request.json.get('content')
        if not content:
            return jsonify({"error": "Content is required"}), 400
        
        # Step 3: Transform the content using the vectorizer
        vectorized_content = vectorizer.transform([content])
        
        # Step 4: Predict tags
        prediction = classifier.predict(vectorized_content)
        predicted_tags = mlb.inverse_transform(prediction)
        
        # Flatten the list of tags
        predicted_tags = [tag for tags in predicted_tags for tag in tags]
        
        # Step 5: Return the predicted tags
        print(f"Input content: {content}")
        print(f"Vectorized content shape: {vectorized_content.shape}")
        print(f"Predicted tags: {predicted_tags}")

        return jsonify({"tags": predicted_tags})
        
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True, host="0.0.0.0", port=5000)
