import json
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField, TextAreaField, BooleanField, widgets, SelectMultipleField
from wtforms.validators import DataRequired, ValidationError


class MultiCheckboxField(SelectMultipleField):
    widget = widgets.ListWidget(prefix_label=False)
    option_widget = widgets.CheckboxInput()


class TestsExecuteForm(FlaskForm):
    flight_records = MultiCheckboxField('Flight Records', choices=[], validators=[DataRequired()])
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    sample_report = BooleanField('Sample Report')
    submit = SubmitField('Run Test')

    def validate_user_config(form, field):
        user_config = json.loads(field.data)
        expected_keys = {'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
        if len(form.flight_records.data) < len(user_config['injection_targets']):
            raise ValidationError(
                'Not enough flight states files provided for each injection_targets.')
